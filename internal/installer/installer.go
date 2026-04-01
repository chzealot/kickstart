package installer

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Tool represents a tool that can be installed.
type Tool struct {
	Name    string
	Check   func() bool
	Install func(dryRun bool) error
}

// IsInstalled checks if a command exists in PATH.
func IsInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// RunCommand executes a command and returns the combined output and error.
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

// DetectPackageManager returns the system package manager on Linux.
func DetectPackageManager() string {
	if runtime.GOOS != "linux" {
		return ""
	}
	managers := []string{"apt-get", "dnf", "yum", "pacman", "zypper", "apk"}
	for _, m := range managers {
		if IsInstalled(m) {
			return m
		}
	}
	return ""
}
