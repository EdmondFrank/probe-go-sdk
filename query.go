package probe

import (
	"fmt"
	"os"
)

// QueryOptions defines options for a query operation.
type QueryOptions struct {
	Path         string   // Path to search in
	Pattern      string   // The ast-grep pattern to search for
	Language     string   // Programming language to search in
	Ignore       []string // Patterns to ignore
	AllowTests   bool     // Include test files
	MaxResults   int      // Maximum number of results
	Format       string   // Output format ('markdown', 'plain', 'json', 'color')
	JSON         bool     // Return results as parsed JSON
}

// Query performs an AST-based pattern match using the probe CLI.
func (c *ProbeClient) Query(opts QueryOptions) (Result, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("Path is required")
	}
	if opts.Pattern == "" {
		return nil, fmt.Errorf("Pattern is required")
	}

	args := []string{"query"}

	// Map struct fields to CLI flags
	if opts.Language != "" {
		args = append(args, "--language", opts.Language)
	}
	for _, ignore := range opts.Ignore {
		args = append(args, "--ignore", ignore)
	}
	if opts.AllowTests {
		args = append(args, "--allow-tests")
	}
	if opts.MaxResults > 0 {
		args = append(args, "--max-results", fmt.Sprintf("%d", opts.MaxResults))
	}
	if opts.Format != "" {
		args = append(args, "--format", opts.Format)
	} else if opts.JSON {
		args = append(args, "--format", "json")
	}

	// Add pattern and path as positional arguments
	args = append(args, opts.Pattern)
	args = append(args, opts.Path)

	// Log the query parameters (for debug parity with npm)
	logMessage := fmt.Sprintf(`Query: pattern=%q path=%q`, opts.Pattern, opts.Path)
	if opts.Language != "" {
		logMessage += fmt.Sprintf(" language=%s", opts.Language)
	}
	if opts.MaxResults > 0 {
		logMessage += fmt.Sprintf(" maxResults=%d", opts.MaxResults)
	}
	if opts.AllowTests {
		logMessage += " allowTests=true"
	}
	fmt.Fprintln(os.Stderr, logMessage)

	return c.runProbeCommand(args...)
}
