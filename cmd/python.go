package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var pythonCmd = &cobra.Command{
	Use:   "python",
	Short: "安装或更新 Python 到最新稳定版",
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

		if cfg.Python == "" {
			ui.Info("配置文件中未定义 python")
			ui.Dim("添加 \"python: latest\" 到配置文件以启用")
			return nil
		}

		ui.Title("Python")
		fmt.Println()

		installPython(cfg.Python, dryRun)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pythonCmd)
}
