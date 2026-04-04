package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/installer"
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
			ui.Warn("配置文件不存在: %s", cfg.Path)
			ui.Dim("请创建 ~/.kickstart/config.yaml，参考 README 中的配置说明")
			return nil
		}

		// Dotfiles
		ui.Section("Dotfiles")
		if cfg.Dotfiles != nil && cfg.Dotfiles.Repo != "" {
			ui.Dim("  repo: %s", cfg.Dotfiles.Repo)
			ui.Dim("  （暂未实现）")
		} else {
			ui.Dim("  未配置")
		}

		// Tools
		ui.Section("安装工具")
		if len(cfg.Tools) == 0 {
			ui.Dim("  未配置")
		} else {
			tools := installer.FromNames(cfg.Tools)
			for _, tool := range tools {
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

		// Repos
		ui.Section("Git 仓库")
		if len(cfg.Repos) == 0 {
			ui.Dim("  未配置")
		} else {
			for _, r := range cfg.Repos {
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

		// Configs
		ui.Section("配置软件")
		if len(cfg.Configs) == 0 {
			ui.Dim("  未配置")
		} else {
			for _, task := range cfg.Configs {
				if dryRun {
					ui.Step("  将执行: %s（dry-run 模式，跳过）", task.Name)
					continue
				}
				ui.Dim("  （暂未实现）%s", task.Name)
			}
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
