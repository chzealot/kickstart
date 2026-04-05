package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/repo"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "执行全部初始化流程（默认）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("kickstart - 一键初始化新电脑环境")
		fmt.Println()

		cfg, err := config.Load(cfgFile)
		if err != nil {
			ui.Error("加载配置失败: %v", err)
			return nil
		}

		if !cfg.Exists() {
			promptInitConfig(cfg)
			return nil
		}

		// Show duplicate warnings from config merging
		for _, w := range config.PopDuplicateWarnings() {
			ui.Warn("  %s", w)
		}

		// Dotfiles
		ui.Section("Dotfiles")
		if cfg.Dotfiles != nil && cfg.Dotfiles.Repo != "" {
			if dryRun {
				ui.Step("  将同步 dotfiles: %s（dry-run 模式，跳过）", cfg.Dotfiles.Repo)
			} else {
				sp := ui.StartSpinner(fmt.Sprintf("  同步 dotfiles: %s ...", cfg.Dotfiles.Repo))
				err := repo.SyncDotfiles(cfg.Dotfiles.Repo)
				sp.Stop()
				if err != nil {
					ui.Error("  dotfiles 同步失败: %v", err)
				} else {
					ui.Success("  dotfiles 已同步")
				}
			}
		} else {
			ui.Dim("  未配置")
		}

		// Go
		ui.Section("Go 语言")
		if cfg.Go == "" {
			ui.Dim("  未配置")
		} else {
			installGo(cfg.Go, dryRun)
		}

		// Python
		ui.Section("Python")
		if cfg.Python == "" {
			ui.Dim("  未配置")
		} else {
			installPython(cfg.Python, dryRun)
		}

		// Tools
		ui.Section("安装工具")
		if len(cfg.Tools) == 0 {
			ui.Dim("  未配置")
		} else if ensurePackageManager(dryRun) {
			installTools(cfg.Tools, dryRun)
		}

		// Repos
		ui.Section("Git 仓库")
		if len(cfg.Repos) == 0 {
			ui.Dim("  未配置")
		} else {
			syncRepos(cfg.Repos, dryRun)
		}

		// Scripts
		ui.Section("执行脚本")
		if len(cfg.Scripts) == 0 {
			ui.Dim("  未配置")
		} else {
			executeScripts(cfg.Scripts, dryRun)
		}

		fmt.Println()
		ui.Success("初始化完成")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Make "run" the default command
	rootCmd.RunE = runCmd.RunE
}
