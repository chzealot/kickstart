package cmd

import (
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "安装工具和软件包",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("安装工具和软件包")
		ui.Dim("暂未配置")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
