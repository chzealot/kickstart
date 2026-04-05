package pyinstall

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// FetchLatestVersion fetches the latest stable Python version string (e.g. "3.14.3").
func FetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	// endoflife.date API returns versions sorted by release date descending
	resp, err := client.Get("https://endoflife.date/api/python.json")
	if err != nil {
		return "", fmt.Errorf("获取 Python 版本信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取 Python 版本信息失败: HTTP %d", resp.StatusCode)
	}

	var cycles []struct {
		Cycle  string `json:"cycle"`
		Latest string `json:"latest"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cycles); err != nil {
		return "", fmt.Errorf("解析 Python 版本信息失败: %w", err)
	}

	if len(cycles) == 0 {
		return "", fmt.Errorf("未找到 Python 版本")
	}

	return cycles[0].Latest, nil
}

// LocalVersion returns the installed Python version (e.g. "3.14.3") or "" if not found.
// Checks python3 command availability.
func LocalVersion() string {
	python3 := findPython3()
	if python3 == "" {
		return ""
	}
	out, err := exec.Command(python3, "--version").Output()
	if err != nil {
		return ""
	}
	// Output: "Python 3.14.3"
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// NeedsInstall returns true if Python needs to be installed or updated.
func NeedsInstall(latest, local string) bool {
	return local == "" || local != latest
}

// Install installs Python using the platform-appropriate method.
// macOS: downloads .pkg from python.org and uses `installer` command.
// Linux: uses system package manager.
func Install(version string) error {
	switch runtime.GOOS {
	case "darwin":
		return installDarwin(version)
	case "linux":
		return installLinux()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// EnsurePythonSymlink creates a `python` symlink pointing to `python3` if missing.
func EnsurePythonSymlink() error {
	python3 := findPython3()
	if python3 == "" {
		return fmt.Errorf("未找到 python3 命令")
	}

	dir := filepath.Dir(python3)
	pythonPath := filepath.Join(dir, "python")

	// Check if python already exists
	if _, err := os.Stat(pythonPath); err == nil {
		return nil // already exists
	}

	// Create symlink (may need sudo for system paths)
	if err := os.Symlink(python3, pythonPath); err != nil {
		// Try with sudo
		cmd := exec.Command("sudo", "ln", "-sf", python3, pythonPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("创建 python 符号链接失败: %w", err)
		}
	}
	return nil
}

// SymlinkStatus returns the python symlink path if it exists, or "" if not.
func SymlinkStatus() string {
	python, err := exec.LookPath("python")
	if err != nil {
		return ""
	}
	return python
}

// --- macOS: install from python.org .pkg ---

func installDarwin(version string) error {
	pkgName := fmt.Sprintf("python-%s-macos11.pkg", version)
	url := fmt.Sprintf("https://www.python.org/ftp/python/%s/%s", version, pkgName)

	// Download .pkg to temp file
	tmpFile := filepath.Join(os.TempDir(), pkgName)
	defer os.Remove(tmpFile)

	if err := downloadFile(url, tmpFile); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// Install using macOS installer command
	cmd := exec.Command("sudo", "installer", "-pkg", tmpFile, "-target", "/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}

	return nil
}

// --- Linux: install via system package manager ---

func installLinux() error {
	pm := detectLinuxPM()
	if pm == "" {
		return fmt.Errorf("未检测到支持的包管理器")
	}

	switch pm {
	case "apt-get":
		if err := runCmd("sudo", "apt-get", "update", "-y"); err != nil {
			return err
		}
		return runCmd("sudo", "apt-get", "install", "-y", "python3")
	case "dnf":
		return runCmd("sudo", "dnf", "install", "-y", "python3")
	case "yum":
		return runCmd("sudo", "yum", "install", "-y", "python3")
	case "pacman":
		return runCmd("sudo", "pacman", "-S", "--noconfirm", "python")
	case "zypper":
		return runCmd("sudo", "zypper", "install", "-y", "python3")
	case "apk":
		return runCmd("sudo", "apk", "add", "python3")
	default:
		return fmt.Errorf("不支持的包管理器: %s", pm)
	}
}

// --- helpers ---

func findPython3() string {
	path, err := exec.LookPath("python3")
	if err != nil {
		return ""
	}
	// Resolve symlinks to get actual path
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

func detectLinuxPM() string {
	managers := []string{"apt-get", "dnf", "yum", "pacman", "zypper", "apk"}
	for _, m := range managers {
		if _, err := exec.LookPath(m); err == nil {
			return m
		}
	}
	return ""
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.ReadFrom(resp.Body)
	return err
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
