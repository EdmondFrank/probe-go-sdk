package probe

import (
	"fmt"
	"os"
)

// SymbolsOptions defines options for a symbols operation.
type SymbolsOptions struct {
	Files      []string // Files to list symbols from
	Cwd        string   // Working directory for resolving relative file paths
	AllowTests bool     // Include test functions/methods
}

// Symbols lists symbols (functions, structs, classes, constants, etc.) in the given files.
// It always returns a parsed JSON array (parity with npm symbols()).
func (c *ProbeClient) Symbols(opts SymbolsOptions) ([]interface{}, error) {
	if len(opts.Files) == 0 {
		return nil, fmt.Errorf("at least one file path is required")
	}

	args := []string{"symbols", "--format", "json"}

	if opts.AllowTests {
		args = append(args, "--allow-tests")
	}

	for _, file := range opts.Files {
		args = append(args, file)
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "\nSymbols: files=%q cwd=%q\n", opts.Files, opts.Cwd)
	}

	// symbols always returns a JSON array; use the array helper
	if opts.Cwd != "" {
		return c.runProbeCommandArrayInDir(opts.Cwd, args...)
	}
	return c.runProbeCommandArray(args...)
}
