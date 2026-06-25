// +build !js

package probe

import (
	"context"
	"os/exec"
	"fmt"
	"os"
	"strings"
	"encoding/json"
	"os/user"
	"path/filepath"
	"runtime"
	"io"
	"net/http"
	"archive/tar"
	"compress/gzip"
	"time"
)

// DefaultTimeout is the default maximum duration for a single probe CLI invocation.
const DefaultTimeout = 5 * time.Minute

// ErrTimeout is returned when a probe CLI command exceeds its timeout.
var ErrTimeout = fmt.Errorf("probe command timed out")

// gracePeriod is the time given to the probe process to shut down gracefully
// (after SIGTERM) before it is force-killed (SIGKILL).
const gracePeriod = 5 * time.Second

// ProbeClient is the main struct for interacting with the probe CLI.
type ProbeClient struct {
	ProbePath string        // Path to the probe CLI binary
	Timeout   time.Duration // Per-command timeout; 0 means no timeout (infinite)
}

// createContext returns a context that is cancelled after the client's timeout
// (if configured). If Timeout is zero or negative, context.Background() is returned.
func (c *ProbeClient) createContext() context.Context {
	if c.Timeout <= 0 {
		return context.Background()
	}
	ctx, _ := context.WithTimeout(context.Background(), c.Timeout)
	return ctx
}

// applyCmdTimeout configures graceful process termination on context cancellation:
// SIGTERM is sent first, and after gracePeriod the process is force-killed (SIGKILL).
func applyCmdTimeout(cmd *exec.Cmd, ctx context.Context) {
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(os.Interrupt)
		}
		return os.ErrProcessDone
	}
	cmd.WaitDelay = gracePeriod
}

// getDefaultProbeBinDir returns the default bin directory for probe
func getDefaultProbeBinDir() string {
	usr, err := user.Current()
	if err != nil {
		return "./bin"
	}
	return filepath.Join(usr.HomeDir, ".probe", "bin")
}

// getDefaultProbePath returns the default path to the probe binary
func getDefaultProbePath() string {
	binDir := getDefaultProbeBinDir()
	binaryName := "probe"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	return filepath.Join(binDir, binaryName)
}

// downloadProbeBinaryIfNeeded downloads the probe binary if not present
func downloadProbeBinaryIfNeeded(binPath string) error {
	if _, err := os.Stat(binPath); err == nil {
		return nil // already exists
	}
	fmt.Fprintf(os.Stderr, "probe binary not found at %s, downloading...\n", binPath)

	// Download from GitHub releases
	owner := "buger"
	repo := "probe"
	osType := runtime.GOOS
	archType := runtime.GOARCH

	var assetName string
	switch osType {
	case "darwin":
		if archType == "arm64" {
			assetName = "probe-darwin-arm64.tar.gz"
		} else {
			assetName = "probe-darwin-amd64.tar.gz"
		}
	case "linux":
		if archType == "arm64" {
			assetName = "probe-linux-arm64.tar.gz"
		} else {
			assetName = "probe-linux-amd64.tar.gz"
		}
	case "windows":
		assetName = "probe-windows-amd64.tar.gz"
	default:
		return fmt.Errorf("unsupported OS/arch: %s/%s", osType, archType)
	}

	// Get latest release info
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()
	type asset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}
	type release struct {
		TagName string  `json:"tag_name"`
		Assets  []asset `json:"assets"`
	}
	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return fmt.Errorf("failed to decode release info: %w", err)
	}
	var downloadURL string
	for _, a := range rel.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("could not find asset %s in latest release", assetName)
	}

	// Download the tar.gz
	tmpFile := binPath + ".tar.gz"
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()
	resp2, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp2.Body.Close()
	if _, err := io.Copy(out, resp2.Body); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	out.Close()

	// Extract the binary
	if err := extractProbeBinary(tmpFile, binPath); err != nil {
		return fmt.Errorf("failed to extract probe binary: %w", err)
	}
	os.Remove(tmpFile)
	return nil
}

// extractProbeBinary extracts the probe binary from a tar.gz archive
func extractProbeBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeReg && (hdr.Name == "probe" || hdr.Name == "probe.exe") {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			if runtime.GOOS != "windows" {
				os.Chmod(destPath, 0o755)
			}
			return nil
		}
	}
	return fmt.Errorf("probe binary not found in archive")
}

/*
NewProbeClient creates a new ProbeClient.

If probePath is empty, it attempts to look up "probe" in the system PATH.
If found, it uses the resolved path; otherwise, it tries to auto-download the binary to ~/.probe/bin/probe.
*/
func NewProbeClient(probePath string) *ProbeClient {
	if probePath == "" {
		// Try system PATH first
		if absPath, err := exec.LookPath("probe"); err == nil {
			probePath = absPath
		} else {
			// Try ~/.probe/bin/probe and auto-download if needed
			probePath = getDefaultProbePath()
			binDir := filepath.Dir(probePath)
			os.MkdirAll(binDir, 0o755)
			if err := downloadProbeBinaryIfNeeded(probePath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to download probe binary: %v\n", err)
			}
		}
	}
	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "probe path: %s\n", probePath)
	}
	return &ProbeClient{ProbePath: probePath, Timeout: DefaultTimeout}
}

// Result is a generic result type for probe operations.
type Result map[string]interface{}

// runProbeCommand runs the probe CLI with the given arguments and returns the parsed JSON output.
// This function mimics the npm/src/cli.js logic: it ensures the probe binary exists and executes it with the provided args.
func (c *ProbeClient) runProbeCommand(args ...string) (Result, error) {
	return c.runProbeCommandInDir("", args...)
}

// runProbeCommandInDir runs the probe CLI in the given working directory (empty = inherit).
func (c *ProbeClient) runProbeCommandInDir(cwd string, args ...string) (Result, error) {
	// Ensure the probe binary exists
	if _, err := exec.LookPath(c.ProbePath); err != nil {
		return nil, fmt.Errorf("probe binary not found at path: %s", c.ProbePath)
	}

	// Prepare the command with timeout context
	ctx := c.createContext()
	cmd := exec.CommandContext(ctx, c.ProbePath, args...)
	applyCmdTimeout(cmd, ctx)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if cwd != "" {
		cmd.Dir = cwd
	}
	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "Executing: %s %s\n", c.ProbePath, strings.Join(args, " "))
	}

	// Capture stdout
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("failed to run probe: %w", err)
	}

	// Try to parse as JSON, fallback to string if not JSON
	var result Result
	if err := json.Unmarshal(out, &result); err != nil {
		// If not JSON, return the raw output as a string in the "output" key
		return Result{"output": strings.TrimSpace(string(out))}, nil
	}
	return result, nil
}

// runProbeCommandWithStdin runs the probe CLI, piping content to stdin, in the given working directory.
func (c *ProbeClient) runProbeCommandWithStdin(content []byte, cwd string, args ...string) (Result, error) {
	// Ensure the probe binary exists
	if _, err := exec.LookPath(c.ProbePath); err != nil {
		return nil, fmt.Errorf("probe binary not found at path: %s", c.ProbePath)
	}

	ctx := c.createContext()
	cmd := exec.CommandContext(ctx, c.ProbePath, args...)
	applyCmdTimeout(cmd, ctx)
	cmd.Stderr = os.Stderr
	if cwd != "" {
		cmd.Dir = cwd
	}
	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "Executing: %s %s\n", c.ProbePath, strings.Join(args, " "))
	}

	// Capture stdout via buffer
	var stdoutBuf strings.Builder
	cmd.Stdout = &stdoutBuf

	// Pipe content to stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("failed to start probe: %w", err)
	}

	// Write content and close stdin
	if _, err := stdin.Write(content); err != nil {
		// If the process was already killed due to timeout, report that
		if ctx.Err() == context.DeadlineExceeded {
			cmd.Wait()
			return nil, ErrTimeout
		}
		cmd.Wait()
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("probe command failed: %w", err)
	}

	out := stdoutBuf.String()

	// Try to parse as JSON, fallback to string
	var result Result
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return Result{"output": strings.TrimSpace(out)}, nil
	}
	return result, nil
}

// runProbeCommandArray runs the probe CLI and parses the output as a JSON array.
// Used for commands like 'symbols' that return a top-level JSON array.
func (c *ProbeClient) runProbeCommandArray(args ...string) ([]interface{}, error) {
	return c.runProbeCommandArrayInDir("", args...)
}

// runProbeCommandArrayInDir runs the probe CLI in the given working directory and parses output as a JSON array.
func (c *ProbeClient) runProbeCommandArrayInDir(cwd string, args ...string) ([]interface{}, error) {
	// Ensure the probe binary exists
	if _, err := exec.LookPath(c.ProbePath); err != nil {
		return nil, fmt.Errorf("probe binary not found at path: %s", c.ProbePath)
	}

	ctx := c.createContext()
	cmd := exec.CommandContext(ctx, c.ProbePath, args...)
	applyCmdTimeout(cmd, ctx)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if cwd != "" {
		cmd.Dir = cwd
	}
	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "Executing: %s %s\n", c.ProbePath, strings.Join(args, " "))
	}

	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("failed to run probe: %w", err)
	}

	var result []interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse probe output as JSON array: %w", err)
	}
	return result, nil
}
