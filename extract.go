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
	AllowTests   bool     // Include test files
	ContextLines int      // Number of context lines to include
	Format       string   // Output format ('markdown', 'plain', 'json')
	JSON         bool     // Return results as parsed JSON
}

// Extract extracts code from files using the probe CLI.
func (c *ProbeClient) Extract(opts ExtractOptions) (Result, error) {
	if (len(opts.Files) == 0 || (len(opts.Files) == 1 && opts.Files[0] == "")) && opts.InputFile == "" {
		return nil, fmt.Errorf("Either Files or InputFile must be provided")
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

	// Add files as positional arguments if provided
	if len(opts.Files) > 0 && !(len(opts.Files) == 1 && opts.Files[0] == "") {
		for _, file := range opts.Files {
			args = append(args, file)
		}
	}

	// Log the extract parameters (for debug parity with npm)
	logMessage := "\nExtract:"
	if len(opts.Files) > 0 && !(len(opts.Files) == 1 && opts.Files[0] == "") {
		logMessage += fmt.Sprintf(" files=\"%s\"", strings.Join(opts.Files, ", "))
	}
	if opts.InputFile != "" {
		logMessage += fmt.Sprintf(" inputFile=\"%s\"", opts.InputFile)
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

	return c.runProbeCommand(args...)
}
