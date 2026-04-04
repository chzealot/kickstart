package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "安装工具和软件包",
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

		if len(cfg.Tools) == 0 {
			ui.Info("配置文件中未定义 tools")
			return nil
		}

		ui.Title("安装工具和软件包")
		fmt.Println()

		// Show duplicate warnings from config merging
		for _, w := range config.PopDuplicateWarnings() {
			ui.Warn("%s", w)
		}

		if !ensurePackageManager(dryRun) {
			return nil
		}

		installTools(cfg.Tools, dryRun)

		fmt.Println()
		ui.Success("所有工具已就绪")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
