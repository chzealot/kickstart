package cmd

import (
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var dotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "管理 dotfiles（symlink 到 $HOME）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("Dotfiles")
		ui.Dim("暂未配置")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dotfilesCmd)
}
