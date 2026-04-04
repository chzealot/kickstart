package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/repo"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var dotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "管理 dotfiles（bare repo 方式部署到 ~/.git）",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			ui.Error("加载配置失败: %v", err)
			return nil
		}

		if !cfg.Exists() {
			promptInitConfig(cfg)
			return nil
		}

		if cfg.Dotfiles == nil || cfg.Dotfiles.Repo == "" {
			ui.Info("配置文件中未定义 dotfiles")
			return nil
		}

		ui.Title("Dotfiles")
		fmt.Println()

		if dryRun {
			ui.Step("将同步 dotfiles: %s（dry-run 模式，跳过）", cfg.Dotfiles.Repo)
			return nil
		}

		sp := ui.StartSpinner(fmt.Sprintf("同步 dotfiles: %s ...", cfg.Dotfiles.Repo))
		err = repo.SyncDotfiles(cfg.Dotfiles.Repo)
		sp.Stop()

		if err != nil {
			ui.Error("dotfiles 同步失败: %v", err)
			return nil
		}

		ui.Success("dotfiles 已同步")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dotfilesCmd)
}
