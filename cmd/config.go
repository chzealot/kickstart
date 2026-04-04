package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "执行软件配置任务",
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

		if len(cfg.Configs) == 0 {
			ui.Info("配置文件中未定义 configs")
			return nil
		}

		ui.Title("配置软件和系统偏好设置")
		fmt.Println()

		executeConfigs(cfg.Configs, dryRun)

		fmt.Println()
		ui.Success("配置任务执行完成")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
