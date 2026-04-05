package cmd

import (
	"fmt"

	"os"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/goinstall"
	"github.com/chzealot/kickstart/internal/installer"
	"github.com/chzealot/kickstart/internal/pyinstall"
	"github.com/chzealot/kickstart/internal/repo"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/chzealot/kickstart/internal/version"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前环境的初始化状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("环境状态")
		fmt.Println()

		ui.Info("版本: %s", version.Version)

		cfg, err := config.Load(cfgFile)
		if err != nil {
			ui.Error("加载配置失败: %v", err)
			return nil
		}

		ui.Info("配置: %s", cfg.Path)
		if !cfg.Exists() {
			ui.Warn("配置文件不存在")
			return nil
		}
		fmt.Println()

		// Dotfiles
		ui.Section("Dotfiles")
		if cfg.Dotfiles != nil && cfg.Dotfiles.Repo != "" {
			ui.Info("  repo: %s", cfg.Dotfiles.Repo)
		} else {
			ui.Dim("  未配置")
		}

		// Go
		ui.Section("Go 语言")
		if cfg.Go != "" {
			local := goinstall.LocalVersion()
			if local != "" {
				ui.Success("  %s ✔", local)
			} else {
				ui.Warn("  未安装")
			}
		} else {
			ui.Dim("  未配置")
		}

		// Python
		ui.Section("Python")
		if cfg.Python != "" {
			local := pyinstall.LocalVersion()
			if local != "" {
				ui.Success("  Python %s ✔", local)
				if sym := pyinstall.SymlinkStatus(); sym != "" {
					ui.Dim("  python → %s", sym)
				}
			} else {
				ui.Warn("  未安装")
			}
		} else {
			ui.Dim("  未配置")
		}

		// Tools
		ui.Section("工具安装")
		if len(cfg.Tools) == 0 {
			ui.Dim("  未配置")
		} else {
			tools := installer.FromNames(cfg.Tools)
			for _, tool := range tools {
				if tool.Check() {
					ui.Success("  %s ✔", tool.Name)
				} else {
					ui.Warn("  %s ✘ 未安装", tool.Name)
				}
			}
		}

		// Repos
		ui.Section("Git 仓库")
		if len(cfg.Repos) == 0 {
			ui.Dim("  未配置")
		} else {
			for _, r := range cfg.Repos {
				path := repo.ExpandHome(r.Path)
				if _, err := os.Stat(path); err == nil {
					ui.Success("  %s → %s ✔", r.URL, r.Path)
				} else {
					ui.Warn("  %s → %s ✘ 未克隆", r.URL, r.Path)
				}
			}
		}

		// Scripts
		ui.Section("配置脚本")
		if len(cfg.Scripts) == 0 {
			ui.Dim("  未配置")
		} else {
			for _, task := range cfg.Scripts {
				ui.Info("  %s", task.Name)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
