package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var confirmUpdate bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "检测环境中可更新的项目（-y 执行更新）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("检测更新")
		fmt.Println()

		// TODO: detect updates for dotfiles, tools, configs
		ui.Dim("暂未配置需要检测的项目")

		if confirmUpdate {
			fmt.Println()
			ui.Info("执行更新...")
			ui.Dim("暂无可更新的项目")
		} else {
			fmt.Println()
			ui.Dim("使用 kickstart update -y 执行更新")
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&confirmUpdate, "yes", "y", false, "确认并执行更新")
	rootCmd.AddCommand(updateCmd)
}
