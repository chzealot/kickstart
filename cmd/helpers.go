package cmd

import (
	"fmt"
	"runtime"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/installer"
	"github.com/chzealot/kickstart/internal/repo"
	"github.com/chzealot/kickstart/internal/runner"
	"github.com/chzealot/kickstart/internal/ui"
)

// ensurePackageManager checks if the platform's package manager is available.
// On macOS, auto-installs Homebrew if missing. On Windows, shows guidance if no PM found.
// Returns true if tools installation can proceed, false if it should be skipped.
func ensurePackageManager(dryRun bool) bool {
	switch runtime.GOOS {
	case "darwin":
		return ensureDarwinPM(dryRun)
	case "windows":
		return ensureWindowsPM(dryRun)
	default:
		return true
	}
}

func ensureDarwinPM(dryRun bool) bool {
	if !installer.NeedsHomebrew() {
		return true
	}

	if dryRun {
		ui.Step("  将安装 Homebrew（dry-run 模式，跳过）")
		return true
	}

	ui.Step("  正在安装 Homebrew...")
	err := installer.InstallHomebrew()
	if err != nil {
		ui.Error("  Homebrew 安装失败: %v", err)
		ui.Dim("  请手动安装: https://brew.sh")
		ui.Dim("  跳过工具安装")
		return false
	}
	ui.Success("  Homebrew 已安装")
	return true
}

func ensureWindowsPM(dryRun bool) bool {
	if !installer.NeedsWindowsPackageManager() {
		return true
	}

	ui.Error("  未检测到 Windows 包管理器")
	ui.Dim("  请安装以下任一工具:")
	ui.Dim("    - winget (Windows 10 1709+ 自带)")
	ui.Dim("    - chocolatey: https://chocolatey.org/install")
	ui.Dim("    - scoop: https://scoop.sh")
	ui.Dim("  跳过工具安装")
	return false
}

// installTools installs tools from the given list using system package managers.
func installTools(tools []string, dryRun bool) {
	for _, tool := range installer.FromNames(tools) {
		if tool.Check() {
			ui.Success("  %s 已安装", tool.Name)
			continue
		}
		if dryRun {
			ui.Step("  将安装 %s（dry-run 模式，跳过）", tool.Name)
			continue
		}
		sp := ui.StartSpinner(fmt.Sprintf("  正在安装 %s...", tool.Name))
		err := tool.Install(false)
		sp.Stop()
		if err != nil {
			ui.Error("  安装 %s 失败: %v", tool.Name, err)
		} else {
			ui.Success("  %s 安装成功", tool.Name)
		}
	}
}

// syncRepos clones or pulls the given list of git repositories.
func syncRepos(repos []config.RepoConfig, dryRun bool) {
	for _, r := range repos {
		if dryRun {
			ui.Step("  将同步 %s → %s（dry-run 模式，跳过）", r.URL, r.Path)
			continue
		}
		sp := ui.StartSpinner(fmt.Sprintf("  同步 %s ...", r.URL))
		err := repo.Sync(r.URL, r.Path)
		sp.Stop()
		if err != nil {
			ui.Error("  %s → %s 失败: %v", r.URL, r.Path, err)
		} else {
			ui.Success("  %s → %s", r.URL, r.Path)
		}
	}
}

// executeConfigs runs the given list of shell configuration tasks.
func executeConfigs(configs []config.ConfigTask, dryRun bool) {
	for _, task := range configs {
		if dryRun {
			ui.Step("  将执行: %s（dry-run 模式，跳过）", task.Name)
			continue
		}
		ui.Step("  执行: %s", task.Name)
		err := runner.RunShell(task.Run)
		if err != nil {
			ui.Error("  %s 失败: %v", task.Name, err)
		} else {
			ui.Success("  %s", task.Name)
		}
	}
}
