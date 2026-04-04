package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/installer"
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

		// Configs
		ui.Section("软件配置")
		if len(cfg.Configs) == 0 {
			ui.Dim("  未配置")
		} else {
			for _, task := range cfg.Configs {
				ui.Info("  %s", task.Name)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
