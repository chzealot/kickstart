package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var goCmd = &cobra.Command{
	Use:   "go",
	Short: "安装或更新 Go 语言到最新稳定版",
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

		if cfg.Go == "" {
			ui.Info("配置文件中未定义 go")
			ui.Dim("添加 \"go: latest\" 到配置文件以启用")
			return nil
		}

		ui.Title("Go 语言")
		fmt.Println()

		installGo(cfg.Go, dryRun)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(goCmd)
}
