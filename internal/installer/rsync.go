package installer

import (
	"fmt"
	"runtime"
)

// Rsync returns the Tool definition for rsync.
func Rsync() Tool {
	return Tool{
		Name:  "rsync",
		Check: func() bool { return IsInstalled("rsync") },
		Install: func(dryRun bool) error {
			return installRsync(dryRun)
		},
	}
}

func installRsync(dryRun bool) error {
	switch runtime.GOOS {
	case "darwin":
		return installRsyncDarwin(dryRun)
	case "linux":
		return installRsyncLinux(dryRun)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func installRsyncDarwin(dryRun bool) error {
	if !IsInstalled("brew") {
		return fmt.Errorf("需要 Homebrew 才能安装 rsync，请先安装 Homebrew: https://brew.sh")
	}
	if dryRun {
		return nil
	}
	return RunCommand("brew", "install", "rsync")
}

func installRsyncLinux(dryRun bool) error {
	pm := DetectPackageManager()
	if pm == "" {
		return fmt.Errorf("未检测到支持的包管理器")
	}
	if dryRun {
		return nil
	}

	switch pm {
	case "apt-get":
		if err := RunCommand("sudo", "apt-get", "update", "-y"); err != nil {
			return err
		}
		return RunCommand("sudo", "apt-get", "install", "-y", "rsync")
	case "dnf":
		return RunCommand("sudo", "dnf", "install", "-y", "rsync")
	case "yum":
		return RunCommand("sudo", "yum", "install", "-y", "rsync")
	case "pacman":
		return RunCommand("sudo", "pacman", "-S", "--noconfirm", "rsync")
	case "zypper":
		return RunCommand("sudo", "zypper", "install", "-y", "rsync")
	case "apk":
		return RunCommand("sudo", "apk", "add", "rsync")
	default:
		return fmt.Errorf("不支持的包管理器: %s", pm)
	}
}
