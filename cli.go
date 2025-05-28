package probe

import (
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
	cmd := exec.Command(c.ProbePath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
