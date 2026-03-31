package cmd

import (
	"fmt"

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
		ui.Info("配置: %s", configPath())
		fmt.Println()

		ui.Dim("Dotfiles    暂未配置")
		ui.Dim("工具安装    暂未配置")
		ui.Dim("软件配置    暂未配置")
		return nil
	},
}

func configPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return "~/.kickstart.yaml（默认）"
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
