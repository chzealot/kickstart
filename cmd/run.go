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
