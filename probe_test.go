package probe

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// skipIfNoProbe skips the test when the probe CLI is not available.
func skipIfNoProbe(t *testing.T) {
	t.Helper()
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration test")
	}
}

// newClient returns a ProbeClient for integration tests.
func newClient() *ProbeClient { return NewProbeClient("") }

// probeHelp returns the output of `probe --help` (cached per test binary run).
var probeHelpOnce struct {
	out  string
	done bool
}

func probeHelp(t *testing.T) string {
	t.Helper()
	if !probeHelpOnce.done {
		c := newClient()
		res, err := c.runProbeCommand("--help")
		if err == nil {
			if v, ok := res["output"]; ok {
				probeHelpOnce.out = v.(string)
			}
		}
		probeHelpOnce.done = true
	}
	return probeHelpOnce.out
}

// probeSubcmdHelp returns the output of `probe <sub> --help`.
func probeSubcmdHelp(t *testing.T, sub string) string {
	t.Helper()
	c := newClient()
	res, err := c.runProbeCommand(sub, "--help")
	if err != nil {
		return ""
	}
	if v, ok := res["output"]; ok {
		return v.(string)
	}
	return ""
}

// skipIfSubcmdUnsupported skips when the probe binary does not list sub as a command.
func skipIfSubcmdUnsupported(t *testing.T, sub string) {
	t.Helper()
	skipIfNoProbe(t)
	if !strings.Contains(probeHelp(t), sub) {
		t.Skipf("probe binary does not support %q subcommand; skipping", sub)
	}
}

// skipIfFlagUnsupported skips when the given flag is not listed in `probe <sub> --help`.
func skipIfFlagUnsupported(t *testing.T, sub, flag string) {
	t.Helper()
	skipIfNoProbe(t)
	if !strings.Contains(probeSubcmdHelp(t, sub), flag) {
		t.Skipf("probe binary does not support %q for %q; skipping", flag, sub)
	}
}

// ---------------------------------------------------------------------------
// Client / binary availability
// ---------------------------------------------------------------------------

func TestProbeAvailable(t *testing.T) {
	// IsProbeAvailable should return a bool without panicking.
	_ = IsProbeAvailable("")
	// With a known-bad path it must return false.
	if IsProbeAvailable("/nonexistent/path/to/probe") {
		t.Error("expected false for non-existent probe path")
	}
}

func TestNewProbeClient_DefaultPath(t *testing.T) {
	client := NewProbeClient("")
	if client.ProbePath == "" {
		t.Error("expected ProbePath to be set")
	}
}

func TestVersion(t *testing.T) {
	skipIfNoProbe(t)
	version, err := newClient().Version()
	if err != nil {
		t.Fatalf("Version() error: %v", err)
	}
	if version == "" {
		t.Fatal("expected non-empty version string")
	}
	// Version should look like "v0.x.y" or "0.x.y".
	if !strings.ContainsAny(version, "0123456789") {
		t.Errorf("version string %q does not contain digits", version)
	}
}

// ---------------------------------------------------------------------------
// Internal helpers – invalid binary path
// ---------------------------------------------------------------------------

func TestRunProbeCommand_InvalidPath(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/path/to/probe"}
	_, err := client.runProbeCommand("--version")
	if err == nil {
		t.Error("expected error for invalid probe path")
	}
}

func TestRunProbeCommandArray_InvalidPath(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/path/to/probe"}
	_, err := client.runProbeCommandArray("symbols", "--format", "json", "probe.go")
	if err == nil {
		t.Error("expected error for invalid probe path")
	}
}

func TestRunProbeCommandWithStdin_InvalidPath(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/path/to/probe"}
	_, err := client.runProbeCommandWithStdin([]byte("probe.go"), "", "extract")
	if err == nil {
		t.Error("expected error for invalid probe path")
	}
}

func TestRunProbeCommand_NonJSONOutput(t *testing.T) {
	skipIfNoProbe(t)
	out, err := newClient().runProbeCommand("--help")
	if err != nil {
		t.Fatalf("runProbeCommand --help: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output for --help")
	}
	raw, ok := out["output"]
	if !ok {
		t.Fatal("expected 'output' key for non-JSON response")
	}
	if !strings.Contains(raw.(string), "Usage") {
		t.Errorf("help output %q does not contain 'Usage'", raw)
	}
}

// ---------------------------------------------------------------------------
// Search – validation (unit, no binary needed)
// ---------------------------------------------------------------------------

func TestSearch_MissingQuery(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/probe"}
	_, err := client.Search(SearchOptions{Path: "."})
	if err == nil || !strings.Contains(err.Error(), "Query is required") {
		t.Errorf("expected 'Query is required' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Search – integration
// ---------------------------------------------------------------------------

func TestSearch(t *testing.T) {
	skipIfNoProbe(t)
	result, err := newClient().Search(SearchOptions{
		Path:       ".",
		Query:      "func",
		MaxResults: 3,
		JSON:       true,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty search result")
	}
}

func TestSearch_ExactAndAllowTests(t *testing.T) {
	skipIfNoProbe(t)
	result, err := newClient().Search(SearchOptions{
		Path:       ".",
		Query:      "func",
		Exact:      true,
		AllowTests: true,
		MaxResults: 2,
		JSON:       true,
	})
	if err != nil {
		t.Fatalf("Search (exact+allowTests) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestSearch_Language(t *testing.T) {
	skipIfNoProbe(t)
	result, err := newClient().Search(SearchOptions{
		Path:       ".",
		Query:      "func",
		Language:   "go",
		MaxResults: 2,
		JSON:       true,
	})
	if err != nil {
		t.Fatalf("Search (language=go) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestSearch_NoMerge(t *testing.T) {
	skipIfNoProbe(t)
	result, err := newClient().Search(SearchOptions{
		Path:       ".",
		Query:      "func",
		NoMerge:    true,
		MaxResults: 2,
		JSON:       true,
	})
	if err != nil {
		t.Fatalf("Search (noMerge) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestSearch_StrictElasticSyntax(t *testing.T) {
	skipIfFlagUnsupported(t, "search", "--strict-elastic-syntax")
	// A valid strict-elastic query uses explicit AND/OR operators.
	result, err := newClient().Search(SearchOptions{
		Path:                ".",
		Query:               "func AND probe",
		StrictElasticSyntax: true,
		MaxResults:          2,
		JSON:                true,
	})
	if err != nil {
		t.Fatalf("Search (strictElasticSyntax) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestSearch_ComplexQuery(t *testing.T) {
	skipIfNoProbe(t)
	if _, err := os.Stat(filepath.Join(".", "probe_test.go")); os.IsNotExist(err) {
		t.Skip("probe_test.go not found; skipping")
	}
	result, err := newClient().Search(SearchOptions{
		Path:  ".",
		Query: "func OR test OR package",
		JSON:  true,
	})
	if err != nil {
		t.Fatalf("Search (complex) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// ---------------------------------------------------------------------------
// Query – validation (unit)
// ---------------------------------------------------------------------------

func TestQuery_MissingPath(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/probe"}
	_, err := client.Query(QueryOptions{Pattern: "func $F() {}"})
	if err == nil || !strings.Contains(err.Error(), "Path is required") {
		t.Errorf("expected 'Path is required' error, got: %v", err)
	}
}

func TestQuery_MissingPattern(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/probe"}
	_, err := client.Query(QueryOptions{Path: "."})
	if err == nil || !strings.Contains(err.Error(), "Pattern is required") {
		t.Errorf("expected 'Pattern is required' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Query – integration
// ---------------------------------------------------------------------------

func TestQuery(t *testing.T) {
	skipIfNoProbe(t)
	result, err := newClient().Query(QueryOptions{
		Path:    ".",
		Pattern: "func $F($A) { $B* }",
		JSON:    true,
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty query result")
	}
}

func TestQuery_IgnoreAndMaxResults(t *testing.T) {
	skipIfNoProbe(t)
	result, err := newClient().Query(QueryOptions{
		Path:       ".",
		Pattern:    "func $F($A) { $B* }",
		Ignore:     []string{"vendor/**"},
		MaxResults: 2,
		JSON:       true,
	})
	if err != nil {
		t.Fatalf("Query (ignore+maxResults) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestQuery_WithContext(t *testing.T) {
	skipIfFlagUnsupported(t, "query", "--with-context")
	result, err := newClient().Query(QueryOptions{
		Path:        ".",
		Pattern:     "func $F($A) { $B* }",
		WithContext: true,
		MaxResults:  2,
		JSON:        true,
	})
	if err != nil {
		t.Fatalf("Query (withContext) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// ---------------------------------------------------------------------------
// Extract – validation (unit)
// ---------------------------------------------------------------------------

func TestExtract_MissingInputs(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/probe"}
	_, err := client.Extract(ExtractOptions{})
	if err == nil {
		t.Error("expected error when no Files/InputFile/Content provided")
	}
	if !strings.Contains(err.Error(), "Extract requires") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExtract_EmptyFilesSlice(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/probe"}
	_, err := client.Extract(ExtractOptions{Files: []string{""}})
	if err == nil {
		t.Error("expected error for empty files slice")
	}
}

// ---------------------------------------------------------------------------
// Extract – integration
// ---------------------------------------------------------------------------

func TestExtract(t *testing.T) {
	skipIfNoProbe(t)
	const filename = "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found; skipping", filename)
	}
	result, err := newClient().Extract(ExtractOptions{
		Files:  []string{filename},
		Format: "json",
		JSON:   true,
	})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty extract result")
	}
}

func TestExtract_InputFile(t *testing.T) {
	skipIfNoProbe(t)
	tmpFile := filepath.Join(t.TempDir(), "extract_input.txt")
	if err := os.WriteFile(tmpFile, []byte("probe_test.go"), 0644); err != nil {
		t.Fatalf("failed to write temp input file: %v", err)
	}
	result, err := newClient().Extract(ExtractOptions{
		InputFile: tmpFile,
		Format:    "json",
		JSON:      true,
	})
	if err != nil {
		t.Fatalf("Extract (inputFile) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty extract result")
	}
}

func TestExtract_Content(t *testing.T) {
	skipIfNoProbe(t)
	const filename = "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found; skipping", filename)
	}
	// Pipe a file path via stdin (Content field).
	result, err := newClient().Extract(ExtractOptions{
		Content: []byte(filename + "\n"),
		JSON:    true,
	})
	if err != nil {
		t.Fatalf("Extract (content/stdin) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty extract result when using Content")
	}
}

func TestExtract_Cwd(t *testing.T) {
	skipIfNoProbe(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	result, err := newClient().Extract(ExtractOptions{
		Files:  []string{"probe_test.go"},
		Cwd:    cwd,
		Format: "json",
		JSON:   true,
	})
	if err != nil {
		t.Fatalf("Extract (cwd) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty extract result")
	}
}

func TestExtract_ContextLinesAndAllowTests(t *testing.T) {
	skipIfNoProbe(t)
	const filename = "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found; skipping", filename)
	}
	result, err := newClient().Extract(ExtractOptions{
		Files:        []string{filename},
		ContextLines: 2,
		AllowTests:   true,
		Format:       "json",
		JSON:         true,
	})
	if err != nil {
		t.Fatalf("Extract (contextLines+allowTests) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty extract result")
	}
}

func TestExtract_LineNumber(t *testing.T) {
	skipIfNoProbe(t)
	// Probe supports "file:line" syntax to extract a specific location.
	result, err := newClient().Extract(ExtractOptions{
		Files:  []string{"probe_test.go:1"},
		Format: "json",
		JSON:   true,
	})
	if err != nil {
		t.Fatalf("Extract (file:line) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result for file:line extraction")
	}
}

// ---------------------------------------------------------------------------
// Symbols – validation (unit)
// ---------------------------------------------------------------------------

func TestSymbols_MissingFiles(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/probe"}
	_, err := client.Symbols(SymbolsOptions{})
	if err == nil || !strings.Contains(err.Error(), "at least one file path is required") {
		t.Errorf("expected 'at least one file path is required' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Symbols – integration
// ---------------------------------------------------------------------------

func TestSymbols(t *testing.T) {
	skipIfSubcmdUnsupported(t, "symbols")
	result, err := newClient().Symbols(SymbolsOptions{
		Files: []string{"probe.go"},
	})
	if err != nil {
		t.Fatalf("Symbols failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected at least one symbol from probe.go")
	}
	// Each element should be a map with at least a "name" or "file" key.
	first, ok := result[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected symbol entry to be a map, got %T", result[0])
	}
	if _, hasFile := first["file"]; !hasFile {
		if _, hasName := first["name"]; !hasName {
			t.Errorf("symbol entry missing both 'file' and 'name' keys: %v", first)
		}
	}
}

func TestSymbols_AllowTests(t *testing.T) {
	skipIfSubcmdUnsupported(t, "symbols")
	result, err := newClient().Symbols(SymbolsOptions{
		Files:      []string{"probe_test.go"},
		AllowTests: true,
	})
	if err != nil {
		t.Fatalf("Symbols (allowTests) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected symbols from probe_test.go with allowTests=true")
	}
}

func TestSymbols_MultipleFiles(t *testing.T) {
	skipIfSubcmdUnsupported(t, "symbols")
	result, err := newClient().Symbols(SymbolsOptions{
		Files: []string{"probe.go", "search.go", "query.go"},
	})
	if err != nil {
		t.Fatalf("Symbols (multiple files) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected symbols from multiple files")
	}
}

func TestSymbols_Cwd(t *testing.T) {
	skipIfSubcmdUnsupported(t, "symbols")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	result, err := newClient().Symbols(SymbolsOptions{
		Files: []string{"probe.go"},
		Cwd:   cwd,
	})
	if err != nil {
		t.Fatalf("Symbols (cwd) failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected symbols when Cwd is set")
	}
}

func TestSymbols_ReturnType(t *testing.T) {
	skipIfSubcmdUnsupported(t, "symbols")
	// Verify the return type is always a slice, even for a single file.
	result, err := newClient().Symbols(SymbolsOptions{
		Files: []string{"symbols.go"},
	})
	if err != nil {
		t.Fatalf("Symbols (return type) failed: %v", err)
	}
	// result must be a []interface{}, length >= 0
	_ = result // already typed as []interface{} by the compiler
}

// ---------------------------------------------------------------------------
// Timeout behavior
// ---------------------------------------------------------------------------

func TestNewProbeClient_DefaultTimeout(t *testing.T) {
	client := NewProbeClient("")
	if client.Timeout != DefaultTimeout {
		t.Errorf("expected default timeout %v, got %v", DefaultTimeout, client.Timeout)
	}
}

func TestRunProbeCommand_TimeoutExceeded(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.runProbeCommand("--help")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestRunProbeCommandWithStdin_TimeoutExceeded(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.runProbeCommandWithStdin([]byte("test"), "", "extract")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestRunProbeCommandArray_TimeoutExceeded(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.runProbeCommandArray("--help")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestVersion_TimeoutExceeded(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.Version()
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestTimeout_ZeroMeansNoTimeout(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = 0 // no timeout
	_, err := client.runProbeCommand("--help")
	if err != nil {
		t.Fatalf("expected no error with Timeout=0, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Timeout behavior – subcommands with real test data
// ---------------------------------------------------------------------------

// common test timeout: long enough for normal execution, short enough to test
const testTimeout = 30 * time.Second

// --- Search ---

func TestSearch_TimeoutExceeded(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.Search(SearchOptions{
		Path:  ".",
		Query: "func",
		JSON:  true,
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestSearch_WithTimeout(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = testTimeout
	result, err := client.Search(SearchOptions{
		Path:       ".",
		Query:      "func",
		MaxResults: 3,
		JSON:       true,
	})
	if err != nil {
		t.Fatalf("Search with timeout failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty search result")
	}
}

// --- Query ---

func TestQuery_TimeoutExceeded(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.Query(QueryOptions{
		Path:    ".",
		Pattern: "func $F($A) { $B* }",
		JSON:    true,
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestQuery_WithTimeout(t *testing.T) {
	skipIfNoProbe(t)
	client := newClient()
	client.Timeout = testTimeout
	result, err := client.Query(QueryOptions{
		Path:    ".",
		Pattern: "func $F($A) { $B* }",
		JSON:    true,
	})
	if err != nil {
		t.Fatalf("Query with timeout failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty query result")
	}
}

// --- Extract (Files) ---

func TestExtract_TimeoutExceeded(t *testing.T) {
	skipIfNoProbe(t)
	const filename = "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found; skipping", filename)
	}
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.Extract(ExtractOptions{
		Files:  []string{filename},
		Format: "json",
		JSON:   true,
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestExtract_WithTimeout(t *testing.T) {
	skipIfNoProbe(t)
	const filename = "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found; skipping", filename)
	}
	client := newClient()
	client.Timeout = testTimeout
	result, err := client.Extract(ExtractOptions{
		Files:  []string{filename},
		Format: "json",
		JSON:   true,
	})
	if err != nil {
		t.Fatalf("Extract with timeout failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty extract result")
	}
}

// --- Extract (Content / stdin) ---

func TestExtract_ContentTimeoutExceeded(t *testing.T) {
	skipIfNoProbe(t)
	const filename = "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found; skipping", filename)
	}
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.Extract(ExtractOptions{
		Content: []byte(filename + "\n"),
		JSON:    true,
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestExtract_ContentWithTimeout(t *testing.T) {
	skipIfNoProbe(t)
	const filename = "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found; skipping", filename)
	}
	client := newClient()
	client.Timeout = testTimeout
	result, err := client.Extract(ExtractOptions{
		Content: []byte(filename + "\n"),
		JSON:    true,
	})
	if err != nil {
		t.Fatalf("Extract (content) with timeout failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty extract result")
	}
}

// --- Symbols ---

func TestSymbols_TimeoutExceeded(t *testing.T) {
	skipIfSubcmdUnsupported(t, "symbols")
	client := newClient()
	client.Timeout = 1 * time.Nanosecond
	_, err := client.Symbols(SymbolsOptions{
		Files: []string{"probe.go"},
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got: %v", err)
	}
}

func TestSymbols_WithTimeout(t *testing.T) {
	skipIfSubcmdUnsupported(t, "symbols")
	client := newClient()
	client.Timeout = testTimeout
	result, err := client.Symbols(SymbolsOptions{
		Files: []string{"probe.go"},
	})
	if err != nil {
		t.Fatalf("Symbols with timeout failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty symbols result")
	}
}
