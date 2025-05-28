package probe

import (
	"os"
	"testing"
	"strings"
)

func TestProbeAvailable(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
}

func TestNewProbeClient_DefaultPath(t *testing.T) {
	client := NewProbeClient("")
	if client.ProbePath == "" {
		t.Error("Expected ProbePath to be set")
	}
}

func TestVersion(t *testing.T) {
	client := NewProbeClient("")
	version, err := client.Version()
	if err != nil {
		t.Errorf("Failed to get probe version: %v", err)
	}
	if version == "" {
		t.Error("Expected non-empty version string")
	}
}

func TestRunProbeCommand_InvalidPath(t *testing.T) {
	client := &ProbeClient{ProbePath: "/nonexistent/path/to/probe"}
	_, err := client.runProbeCommand("--version")
	if err == nil {
		t.Error("Expected error for invalid probe path")
	}
}

func TestSearch(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
	client := NewProbeClient("")
	opts := SearchOptions{
		Path:       ".",
		Query:      "func",
		MaxResults: 1,
		JSON:       true,
	}
	result, err := client.Search(opts)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil search result")
	}
}

func TestQuery(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
	client := NewProbeClient("")
	opts := QueryOptions{
		Path:    ".",
		Pattern: "func $F($A) { $B* }",
		JSON:    true,
	}
	result, err := client.Query(opts)
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil query result")
	}
}

func TestExtract(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
	client := NewProbeClient("")
	// Try to extract from this test file itself
	filename := "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("Test file %s not found; skipping extract test", filename)
	}
	opts := ExtractOptions{
		Files:  []string{filename},
		Format: "json",
		JSON:   true,
	}
	result, err := client.Extract(opts)
	if err != nil {
		t.Errorf("Extract failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil extract result")
	}
}

func TestExtract_InputFile(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
	client := NewProbeClient("")
	// Write a temp file with a file path
	tmpFile := "tmp_extract_input.txt"
	content := "probe_test.go"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp input file: %v", err)
	}
	defer os.Remove(tmpFile)
	opts := ExtractOptions{
		InputFile: tmpFile,
		Format:    "json",
		JSON:      true,
	}
	result, err := client.Extract(opts)
	if err != nil {
		t.Errorf("Extract with input file failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil extract result")
	}
}

func TestSearch_ExactAndAllowTests(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
	client := NewProbeClient("")
	opts := SearchOptions{
		Path:       ".",
		Query:      "func",
		Exact:      true,
		AllowTests: true,
		MaxResults: 1,
		JSON:       true,
	}
	result, err := client.Search(opts)
	if err != nil {
		t.Errorf("Search with exact and allowTests failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil search result")
	}
}

func TestQuery_IgnoreAndMaxResults(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
	client := NewProbeClient("")
	opts := QueryOptions{
		Path:       ".",
		Pattern:    "func $F($A) { $B* }",
		Ignore:     []string{"vendor/**"},
		MaxResults: 2,
		JSON:       true,
	}
	result, err := client.Query(opts)
	if err != nil {
		t.Errorf("Query with ignore and maxResults failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil query result")
	}
}

func TestExtract_ContextLinesAndAllowTests(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
	client := NewProbeClient("")
	filename := "probe_test.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("Test file %s not found; skipping extract test", filename)
	}
	opts := ExtractOptions{
		Files:        []string{filename},
		ContextLines: 2,
		AllowTests:   true,
		Format:       "json",
		JSON:         true,
	}
	result, err := client.Extract(opts)
	if err != nil {
		t.Errorf("Extract with contextLines and allowTests failed: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil extract result")
	}
}

func TestRunProbeCommand_NonJSONOutput(t *testing.T) {
	if !IsProbeAvailable("") {
		t.Skip("probe CLI not found in PATH; skipping integration tests")
	}
	client := NewProbeClient("")
	// Use a command that will output plain text, not JSON
	out, err := client.runProbeCommand("--help")
	if err != nil {
		t.Errorf("runProbeCommand --help failed: %v", err)
	}
	if out == nil {
		t.Error("Expected non-nil output for --help")
	}
	if _, ok := out["output"]; !ok {
		t.Error("Expected 'output' key in result for non-JSON output")
	}
	if !strings.Contains(out["output"].(string), "Usage") {
		t.Error("Expected help output to contain 'Usage'")
	}
}
