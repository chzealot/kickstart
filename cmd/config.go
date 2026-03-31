package cmd

import (
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置软件和系统偏好设置",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("配置软件和系统偏好设置")
		ui.Dim("暂未配置")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
