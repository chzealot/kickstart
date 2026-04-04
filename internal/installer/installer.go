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

// FromNames creates a list of Tools from tool name strings.
// Each tool uses the system package manager for installation.
func FromNames(names []string) []Tool {
	tools := make([]Tool, len(names))
	for i, name := range names {
		tools[i] = newGenericTool(name)
	}
	return tools
}

func newGenericTool(name string) Tool {
	return Tool{
		Name:  name,
		Check: func() bool { return IsInstalled(name) },
		Install: func(dryRun bool) error {
			if dryRun {
				return nil
			}
			return installWithPackageManager(name)
		},
	}
}

func installWithPackageManager(name string) error {
	switch runtime.GOOS {
	case "darwin":
		if !IsInstalled("brew") {
			return fmt.Errorf("需要 Homebrew 才能安装 %s，请先安装: https://brew.sh", name)
		}
		return RunCommand("brew", "install", name)
	case "linux":
		return installWithLinuxPM(name)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func installWithLinuxPM(name string) error {
	pm := DetectPackageManager()
	if pm == "" {
		return fmt.Errorf("未检测到支持的包管理器")
	}

	switch pm {
	case "apt-get":
		if err := RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
			return err
		}
		return RunCommand("sudo", "apt-get", "install", "-y", name)
	case "dnf":
		return RunCommand("sudo", "dnf", "install", "-y", name)
	case "yum":
		return RunCommand("sudo", "yum", "install", "-y", name)
	case "pacman":
		return RunCommand("sudo", "pacman", "-S", "--noconfirm", name)
	case "zypper":
		return RunCommand("sudo", "zypper", "install", "-y", name)
	case "apk":
		return RunCommand("sudo", "apk", "add", name)
	default:
		return fmt.Errorf("不支持的包管理器: %s", pm)
	}
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
