# Probe Go SDK

Probe is an AI-native code intelligence system designed to bridge the gap between human developers, AI assistants, and complex codebases.

This Go SDK provides semantic code search, intelligent code extraction, and AST-based pattern matching capabilities, making large codebases more understandable and accessible.

## Features

- **Semantic Code Search**: Find code by meaning, not just text.
- **AST-based Pattern Matching**: Query code structure using patterns.
- **Intelligent Code Extraction**: Extract relevant code snippets with context.
- **CLI Integration**: Wraps the [probe CLI](https://github.com/buger/probe) for seamless use in Go projects.

## Installation

```bash
go get github.com/edmondfrank/probe-go-sdk
```

## Requirements

- Go 1.21 or newer
- The [probe CLI](https://github.com/buger/probe) (auto-downloaded if not found)

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/edmondfrank/probe-go-sdk"
)

func main() {
    client := probe.NewProbeClient("")
    opts := probe.SearchOptions{
        Path:  ".",
        Query: "func",
        JSON:  true,
    }
    result, err := client.Search(opts)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%#v\n", result)
}
```

### CLI Example

A simple CLI is provided in `cli/main.go`:

```bash
cd cli
go run main.go --mode search --query "func"
```

## API

- `NewProbeClient(probePath string) *ProbeClient`: Create a new client. If `probePath` is empty, the SDK will auto-detect or download the probe binary.
- `Search(opts SearchOptions) (Result, error)`: Semantic code search.
- `Query(opts QueryOptions) (Result, error)`: AST-based pattern matching.
- `Extract(opts ExtractOptions) (Result, error)`: Extract code snippets.

See GoDoc comments in the source for detailed option fields.

## Development

- Run tests:
  ```bash
  go test
  ```
- Lint:
  ```bash
  golint ./...
  ```

## License

MIT

## Acknowledgements

- [probe CLI](https://github.com/buger/probe) by Ivan Bugaychenko
- Inspired by modern code intelligence and AI developer tools
