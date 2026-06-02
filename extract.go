package probe

import (
	"fmt"
	"os"
	"strings"
)

// ExtractOptions defines options for an extract operation.
type ExtractOptions struct {
	Files        []string // Files to extract from (can include line numbers with colon)
	InputFile    string   // Path to a file containing unstructured text to extract file paths from
	Content      []byte   // Content to pipe to stdin (e.g. git diff output); alternative to InputFile
	Cwd          string   // Working directory for resolving relative file paths
	AllowTests   bool     // Include test files
	ContextLines int      // Number of context lines to include
	Format       string   // Output format ('markdown', 'plain', 'json')
	LSP          bool     // --lsp: use LSP for call hierarchy and reference graphs
	JSON         bool     // Return results as parsed JSON
}

// Extract extracts code from files using the probe CLI.
func (c *ProbeClient) Extract(opts ExtractOptions) (Result, error) {
	hasFiles := len(opts.Files) > 0 && !(len(opts.Files) == 1 && opts.Files[0] == "")
	hasInputFile := opts.InputFile != ""
	hasContent := len(opts.Content) > 0

	if !hasFiles && !hasInputFile && !hasContent {
		return nil, fmt.Errorf("Extract requires one of: Files (array of file paths), InputFile (path to input file), or Content (bytes for stdin)")
	}

	args := []string{"extract"}

	// Map struct fields to CLI flags
	if opts.AllowTests {
		args = append(args, "--allow-tests")
	}
	if opts.ContextLines > 0 {
		args = append(args, "--context", fmt.Sprintf("%d", opts.ContextLines))
	}
	if opts.Format != "" {
		args = append(args, "--format", opts.Format)
	} else if opts.JSON {
		args = append(args, "--format", "json")
	}
	if opts.InputFile != "" {
		args = append(args, "--input-file", opts.InputFile)
	}
	if opts.LSP {
		args = append(args, "--lsp")
	}

	// Add files as positional arguments if provided
	if hasFiles {
		for _, file := range opts.Files {
			args = append(args, file)
		}
	}

	// Log the extract parameters only when DEBUG=1 (parity with npm)
	if os.Getenv("DEBUG") == "1" {
		logMessage := "\nExtract:"
		if hasFiles {
			logMessage += fmt.Sprintf(" files=%q", strings.Join(opts.Files, ", "))
		}
		if opts.InputFile != "" {
			logMessage += fmt.Sprintf(" inputFile=%q", opts.InputFile)
		}
		if hasContent {
			logMessage += fmt.Sprintf(" content=(%d bytes)", len(opts.Content))
		}
		if opts.Cwd != "" {
			logMessage += fmt.Sprintf(" cwd=%q", opts.Cwd)
		}
		if opts.AllowTests {
			logMessage += " allowTests=true"
		}
		if opts.ContextLines > 0 {
			logMessage += fmt.Sprintf(" contextLines=%d", opts.ContextLines)
		}
		if opts.Format != "" {
			logMessage += fmt.Sprintf(" format=%s", opts.Format)
		}
		if opts.JSON {
			logMessage += " json=true"
		}
		fmt.Fprintln(os.Stderr, logMessage)
	}

	// If content is provided, pipe it to stdin
	if hasContent {
		return c.runProbeCommandWithStdin(opts.Content, opts.Cwd, args...)
	}

	return c.runProbeCommandInDir(opts.Cwd, args...)
}
