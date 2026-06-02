# Probe Go SDK

A Go SDK for [probe](https://github.com/buger/probe) — an AI-native code intelligence tool. Wraps the probe CLI to provide semantic search, AST-based pattern matching, code extraction, and symbol listing from Go programs.

## Features

- **Semantic Search** — find code by meaning across a directory tree
- **AST Pattern Matching** — query code structure with ast-grep patterns
- **Code Extraction** — pull exact code blocks from files, line ranges, or stdin
- **Symbol Listing** — enumerate functions, types, and constants in files
- **Auto-download** — fetches the probe binary automatically if not found on `$PATH`
- **Debug logging** — set `DEBUG=1` for verbose probe path and command tracing

## Installation

```bash
go get github.com/edmondfrank/probe-go-sdk
```

**Requirements:** Go 1.21+. The probe CLI binary is auto-downloaded to `~/.probe/bin/` on first use if it is not already on `$PATH`.

## Quick Start

```go
package main

import (
    "fmt"
    probe "github.com/edmondfrank/probe-go-sdk"
)

func main() {
    client := probe.NewProbeClient("") // resolves or downloads probe automatically

    result, err := client.Search(probe.SearchOptions{
        Path:  ".",
        Query: "http handler",
        JSON:  true,
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(result)
}
```

## API

### `NewProbeClient(probePath string) *ProbeClient`

Creates a client. When `probePath` is empty the SDK:
1. Looks for `probe` on `$PATH`.
2. Falls back to `~/.probe/bin/probe`, downloading the latest release if absent.

---

### `Search(opts SearchOptions) (Result, error)`

Semantic code search. `Result` is `map[string]interface{}`.

```go
result, err := client.Search(probe.SearchOptions{
    Path:                ".",          // directory to search (default ".")
    Query:               "error handling",
    Language:            "go",         // limit to a language
    Exact:               false,        // exact match (no tokenisation)
    StrictElasticSyntax: false,        // require explicit AND/OR operators
    AllowTests:          false,        // include test files
    MaxResults:          20,
    MaxTokens:           20000,        // default 20000
    MaxBytes:            0,
    NoMerge:             false,
    MergeThreshold:      0,
    Reranker:            "bm25",       // bm25 | hybrid | hybrid2 | tfidf
    FrequencySearch:     false,
    FilesOnly:           false,
    ExcludeFilenames:    false,
    Session:             "",           // also read from $PROBE_SESSION_ID
    Timeout:             30,           // seconds, default 30
    LSP:                 false,
    DryRun:              false,
    Format:              "",           // markdown | plain | json | color
    JSON:                true,         // parse output as JSON
})
```

---

### `Query(opts QueryOptions) (Result, error)`

AST-based structural search using [ast-grep](https://ast-grep.github.io/) patterns.

```go
result, err := client.Query(probe.QueryOptions{
    Path:        ".",
    Pattern:     "func $F($A) error { $B* }",
    Language:    "go",
    Ignore:      []string{"vendor/**"},
    AllowTests:  false,
    WithContext:  false,   // include owning source-block context in JSON
    MaxResults:  10,
    Format:      "",       // markdown | plain | json | color
    JSON:        true,
})
```

---

### `Extract(opts ExtractOptions) (Result, error)`

Extracts code blocks from specific files or line ranges. Supports three input modes — use exactly one:

| Field | Description |
|---|---|
| `Files []string` | File paths, optionally with line numbers: `"main.go:42"` |
| `InputFile string` | Path to a file whose text is scanned for file references |
| `Content []byte` | Raw bytes piped to probe's stdin (e.g. a `git diff`) |

```go
// From explicit files
result, err := client.Extract(probe.ExtractOptions{
    Files:        []string{"main.go", "handler.go:120"},
    Cwd:          "/path/to/project",  // resolve relative paths from here
    ContextLines: 3,
    AllowTests:   false,
    LSP:          false,               // use LSP for call hierarchy
    Format:       "json",
    JSON:         true,
})

// From stdin content (e.g. a git diff)
diff, _ := os.ReadFile("my.diff")
result, err := client.Extract(probe.ExtractOptions{
    Content: diff,
    JSON:    true,
})
```

---

### `Symbols(opts SymbolsOptions) ([]interface{}, error)`

Lists symbols (functions, types, constants, …) in one or more files. Always returns a parsed JSON array.

> **Note:** requires probe ≥ v0.7. The SDK skips this call gracefully when the binary is older.

```go
symbols, err := client.Symbols(probe.SymbolsOptions{
    Files:      []string{"main.go", "server.go"},
    Cwd:        "/path/to/project",
    AllowTests: false,
})
for _, s := range symbols {
    fmt.Println(s)
}
```

---

### `Version() (string, error)`

Returns the probe binary version string (e.g. `"probe-code 0.6.0"`).

```go
v, err := client.Version()
```

---

### `IsProbeAvailable(probePath string) bool`

Returns `true` if the probe binary can be found at `probePath` (or on `$PATH` when empty).

## Debug Logging

Set `DEBUG=1` to print diagnostic output to stderr:

```
probe path: /usr/local/bin/probe
Search: query="http handler" path="." maxTokens=20000 timeout=30
Executing: /usr/local/bin/probe search --format json --max-tokens 20000 --timeout 30 http handler .
```

The following events are logged:

| Event | Trigger |
|---|---|
| `probe path: …` | `NewProbeClient` — resolved binary path |
| `Search: …` | `Search()` — query parameters |
| `Query: …` | `Query()` — pattern and path |
| `Extract: …` | `Extract()` — files / inputFile / content size |
| `Symbols: …` | `Symbols()` — files and cwd |
| `Executing: …` | Every probe invocation — full command line |

## CLI Demo

A minimal CLI wrapper lives in `cli/`:

```bash
cd cli
go run main.go --mode search --path . --query "func"
go run main.go --mode query  --path . --pattern "func $F($A) { $B* }"
go run main.go --mode extract --files "main.go:1"
```

## Development

```bash
# Run all tests (integration tests skip when probe is not on $PATH)
go test ./...

# Run with debug output
DEBUG=1 go test ./... -v -run TestSearch

# Build
go build ./...
```

## License

MIT

## Acknowledgements

- [probe](https://github.com/buger/probe) by Ivan Bugaychenko
- Inspired by modern code intelligence and AI developer tooling
