package probe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsProbeAvailable checks if the probe CLI is available in PATH or at the given path.
func IsProbeAvailable(probePath string) bool {
	if probePath == "" {
		probePath = "probe"
	}
	_, err := exec.LookPath(probePath)
	return err == nil
}

// Version returns the version of the probe CLI.
func (c *ProbeClient) Version() (string, error) {
	ctx := c.createContext()
	cmd := exec.CommandContext(ctx, c.ProbePath, "--version")
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(os.Interrupt)
		}
		return os.ErrProcessDone
	}
	cmd.WaitDelay = gracePeriod

	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", ErrTimeout
		}
		return "", fmt.Errorf("failed to get probe version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
