package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edmondfrank/probe-go-sdk"
)

func main() {
	// Simple CLI demo for probe-go-sdk
	var (
		mode       string
		path       string
		query      string
		pattern    string
		files      string
		inputFile  string
		language   string
		format     string
		maxResults int
		context    int
		allowTests bool
		exact      bool
		jsonOut    bool
		timeout    time.Duration
	)

	flag.StringVar(&mode, "mode", "search", "Mode: search | query | extract | symbols")
	flag.StringVar(&path, "path", ".", "Path to search in")
	flag.StringVar(&query, "query", "", "Search query (for search mode)")
	flag.StringVar(&pattern, "pattern", "", "Pattern (for query mode)")
	flag.StringVar(&files, "files", "", "Comma-separated list of files (for extract/symbols mode)")
	flag.StringVar(&inputFile, "input-file", "", "Input file for extract mode")
	flag.StringVar(&language, "language", "", "Programming language")
	flag.StringVar(&format, "format", "json", "Output format (json, markdown, plain)")
	flag.IntVar(&maxResults, "max-results", 5, "Maximum number of results")
	flag.IntVar(&context, "context", 0, "Context lines (extract mode)")
	flag.BoolVar(&allowTests, "allow-tests", false, "Include test files")
	flag.BoolVar(&exact, "exact", false, "Exact match (search mode)")
	flag.BoolVar(&jsonOut, "json", true, "Output as JSON")
	flag.DurationVar(&timeout, "timeout", 0, "Per-command timeout (0 = default 5m)")

	flag.Parse()

	client := probe.NewProbeClient("")
	if timeout > 0 {
		client.Timeout = timeout
	}

	switch mode {
	case "search":
		opts := probe.SearchOptions{
			Path:       path,
			Query:      query,
			Language:   language,
			MaxResults: maxResults,
			AllowTests: allowTests,
			Exact:      exact,
			Format:     format,
			JSON:       jsonOut,
		}
		result, err := client.Search(opts)
		if err != nil {
			handleErr("Search", err)
		}
		printResult(result, jsonOut)
	case "query":
		opts := probe.QueryOptions{
			Path:       path,
			Pattern:    pattern,
			Language:   language,
			MaxResults: maxResults,
			AllowTests: allowTests,
			Format:     format,
			JSON:       jsonOut,
		}
		result, err := client.Query(opts)
		if err != nil {
			handleErr("Query", err)
		}
		printResult(result, jsonOut)
	case "extract":
		var fileList []string
		if files != "" {
			for _, f := range strings.Split(files, ",") {
				trimmed := strings.TrimSpace(f)
				if trimmed != "" {
					fileList = append(fileList, trimmed)
				}
			}
		}
		opts := probe.ExtractOptions{
			Files:        fileList,
			InputFile:    inputFile,
			AllowTests:   allowTests,
			ContextLines: context,
			Format:       format,
			JSON:         jsonOut,
		}
		result, err := client.Extract(opts)
		if err != nil {
			handleErr("Extract", err)
		}
		printResult(result, jsonOut)
	case "symbols":
		var fileList []string
		if files != "" {
			for _, f := range strings.Split(files, ",") {
				trimmed := strings.TrimSpace(f)
				if trimmed != "" {
					fileList = append(fileList, trimmed)
				}
			}
		}
		result, err := client.Symbols(probe.SymbolsOptions{
			Files:      fileList,
			AllowTests: allowTests,
		})
		if err != nil {
			handleErr("Symbols", err)
		}
		// Symbols returns an array, wrap for uniform printing
		printArrayResult(result, jsonOut)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

func printResult(result probe.Result, jsonOut bool) {
	if jsonOut {
		// Pretty print JSON
		b, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("%#v\n", result)
		} else {
			fmt.Println(string(b))
		}
	} else {
		// Print as string or fallback
		if out, ok := result["output"]; ok {
			fmt.Println(out)
		} else {
			fmt.Printf("%#v\n", result)
		}
	}
}

func printArrayResult(result []interface{}, jsonOut bool) {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("%#v\n", result)
	} else {
		fmt.Println(string(b))
	}
}

// handleErr prints a user-friendly error message and exits.
// Timeout errors are reported distinctly so callers can distinguish
// resource-exhaustion situations from ordinary failures.
func handleErr(mode string, err error) {
	if errors.Is(err, probe.ErrTimeout) {
		fmt.Fprintf(os.Stderr, "%s timed out and was terminated\n", mode)
	} else {
		fmt.Fprintf(os.Stderr, "%s error: %v\n", mode, err)
	}
	os.Exit(1)
}
