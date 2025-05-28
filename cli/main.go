package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

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
	)

	flag.StringVar(&mode, "mode", "search", "Mode: search | query | extract")
	flag.StringVar(&path, "path", ".", "Path to search in")
	flag.StringVar(&query, "query", "", "Search query (for search mode)")
	flag.StringVar(&pattern, "pattern", "", "Pattern (for query mode)")
	flag.StringVar(&files, "files", "", "Comma-separated list of files (for extract mode)")
	flag.StringVar(&inputFile, "input-file", "", "Input file for extract mode")
	flag.StringVar(&language, "language", "", "Programming language")
	flag.StringVar(&format, "format", "json", "Output format (json, markdown, plain)")
	flag.IntVar(&maxResults, "max-results", 5, "Maximum number of results")
	flag.IntVar(&context, "context", 0, "Context lines (extract mode)")
	flag.BoolVar(&allowTests, "allow-tests", false, "Include test files")
	flag.BoolVar(&exact, "exact", false, "Exact match (search mode)")
	flag.BoolVar(&jsonOut, "json", true, "Output as JSON")

	flag.Parse()

	client := probe.NewProbeClient("")

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
			fmt.Fprintf(os.Stderr, "Search error: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Extract error: %v\n", err)
			os.Exit(1)
		}
		printResult(result, jsonOut)
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
