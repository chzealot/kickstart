package installer

import (
	"fmt"
	"os"
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
			if err := InstallHomebrew(); err != nil {
				return fmt.Errorf("自动安装 Homebrew 失败: %w\n请手动安装: https://brew.sh", err)
			}
		}
		return RunCommand("brew", "install", name)
	case "linux":
		return installWithLinuxPM(name)
	case "windows":
		return installWithWindowsPM(name)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// NeedsHomebrew returns true if on macOS and brew is not installed.
func NeedsHomebrew() bool {
	return runtime.GOOS == "darwin" && !IsInstalled("brew")
}

// InstallHomebrew installs Homebrew on macOS using the official install script.
// After installation, it adds the brew binary to PATH for the current process.
func InstallHomebrew() error {
	cmd := exec.Command("/bin/bash", "-c",
		`NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Add brew to PATH for current process
	// Apple Silicon: /opt/homebrew/bin, Intel: /usr/local/bin
	brewPaths := []string{"/opt/homebrew/bin", "/usr/local/bin"}
	for _, p := range brewPaths {
		brewBin := p + "/brew"
		if _, err := os.Stat(brewBin); err == nil {
			os.Setenv("PATH", p+":"+os.Getenv("PATH"))
			break
		}
	}

	if !IsInstalled("brew") {
		return fmt.Errorf("安装完成但未找到 brew 命令")
	}

	return nil
}

// NeedsWindowsPackageManager returns true if on Windows and no package manager is found.
func NeedsWindowsPackageManager() bool {
	return runtime.GOOS == "windows" && detectWindowsPM() == ""
}

func detectWindowsPM() string {
	managers := []string{"winget", "choco", "scoop"}
	for _, m := range managers {
		if IsInstalled(m) {
			return m
		}
	}
	return ""
}

func installWithWindowsPM(name string) error {
	pm := detectWindowsPM()
	if pm == "" {
		return fmt.Errorf("未检测到 Windows 包管理器\n请安装以下任一工具:\n  - winget (Windows 10 1709+ 自带)\n  - chocolatey: https://chocolatey.org/install\n  - scoop: https://scoop.sh")
	}

	switch pm {
	case "winget":
		return RunCommand("winget", "install", "--accept-source-agreements", "--accept-package-agreements", "-e", "--id", name)
	case "choco":
		return RunCommand("choco", "install", name, "-y")
	case "scoop":
		return RunCommand("scoop", "install", name)
	default:
		return fmt.Errorf("不支持的包管理器: %s", pm)
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
