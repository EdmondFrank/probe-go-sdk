package probe

import (
	"fmt"
	"os"
)

// SearchOptions defines options for a search operation.
type SearchOptions struct {
	Path                string   // Path to search in (like npm: path)
	Query               string   // Search query or pattern (like npm: query)
	FilesOnly           bool     // --files-only
	Ignore              []string // --ignore
	ExcludeFilenames    bool     // --exclude-filenames
	Reranker            string   // --reranker
	FrequencySearch     bool     // --frequency
	Exact               bool     // --exact
	StrictElasticSyntax bool     // --strict-elastic-syntax
	MaxResults          int      // --max-results
	MaxBytes            int      // --max-bytes
	MaxTokens           int      // --max-tokens
	AllowTests          bool     // --allow-tests
	NoMerge             bool     // --no-merge
	MergeThreshold      int      // --merge-threshold
	Session             string   // --session
	Timeout             int      // --timeout
	Language            string   // --language
	Format              string   // --format
	LSP                 bool     // --lsp
	DryRun              bool     // --dry-run
	BinaryOptions       map[string]interface{} // not used in Go, but for parity
	JSON                bool     // Return results as parsed JSON
}

// Search performs a semantic code search using the probe CLI.
func (c *ProbeClient) Search(opts SearchOptions) (Result, error) {
	if opts.Path == "" {
		opts.Path = "."
	}
	if opts.Query == "" {
		return nil, fmt.Errorf("Query is required")
	}

	args := []string{"search"}

	// Map struct fields to CLI flags
	if opts.FilesOnly {
		args = append(args, "--files-only")
	}
	for _, ignore := range opts.Ignore {
		args = append(args, "--ignore", ignore)
	}
	if opts.ExcludeFilenames {
		args = append(args, "--exclude-filenames")
	}
	if opts.Reranker != "" {
		args = append(args, "--reranker", opts.Reranker)
	}
	if opts.FrequencySearch {
		args = append(args, "--frequency")
	}
	if opts.Exact {
		args = append(args, "--exact")
	}
	if opts.StrictElasticSyntax {
		args = append(args, "--strict-elastic-syntax")
	}
	if opts.MaxResults > 0 {
		args = append(args, "--max-results", fmt.Sprintf("%d", opts.MaxResults))
	}
	if opts.MaxBytes > 0 {
		args = append(args, "--max-bytes", fmt.Sprintf("%d", opts.MaxBytes))
	}
	if opts.MaxTokens > 0 {
		args = append(args, "--max-tokens", fmt.Sprintf("%d", opts.MaxTokens))
	} else {
		args = append(args, "--max-tokens", "20000")
	}
	if opts.AllowTests {
		args = append(args, "--allow-tests")
	}
	if opts.NoMerge {
		args = append(args, "--no-merge")
	}
	if opts.MergeThreshold > 0 {
		args = append(args, "--merge-threshold", fmt.Sprintf("%d", opts.MergeThreshold))
	}
	if opts.Session != "" {
		args = append(args, "--session", opts.Session)
	} else if session := os.Getenv("PROBE_SESSION_ID"); session != "" {
		args = append(args, "--session", session)
	}
	if opts.Timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%d", opts.Timeout))
	} else {
		args = append(args, "--timeout", "30")
	}
	if opts.Language != "" {
		args = append(args, "--language", opts.Language)
	}
	if opts.Format != "" {
		args = append(args, "--format", opts.Format)
	} else if opts.JSON {
		args = append(args, "--format", "json")
	}
	if opts.LSP {
		args = append(args, "--lsp")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	// Add query and path as positional arguments
	args = append(args, opts.Query)
	args = append(args, opts.Path)

	// Log the search parameters only when DEBUG=1 (parity with npm)
	if os.Getenv("DEBUG") == "1" {
		logMessage := fmt.Sprintf("\nSearch: query=%q path=%q", opts.Query, opts.Path)
		if opts.MaxResults > 0 {
			logMessage += fmt.Sprintf(" maxResults=%d", opts.MaxResults)
		}
		logMessage += fmt.Sprintf(" maxTokens=%d", opts.MaxTokens)
		logMessage += fmt.Sprintf(" timeout=%d", opts.Timeout)
		if opts.AllowTests {
			logMessage += " allowTests=true"
		}
		if opts.Language != "" {
			logMessage += fmt.Sprintf(" language=%s", opts.Language)
		}
		if opts.Exact {
			logMessage += " exact=true"
		}
		if opts.Session != "" {
			logMessage += fmt.Sprintf(" session=%s", opts.Session)
		}
		fmt.Fprintln(os.Stderr, logMessage)
	}

	return c.runProbeCommand(args...)
}
