package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "执行全部初始化流程（默认）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("kickstart - 一键初始化新电脑环境")
		fmt.Println()

		ui.Section("Dotfiles")
		ui.Dim("  暂未配置")

		ui.Section("安装工具")
		ui.Dim("  暂未配置")

		ui.Section("配置软件")
		ui.Dim("  暂未配置")

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
