package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "同步 Git 仓库（clone 或 pull）",
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

		if len(cfg.Repos) == 0 {
			ui.Info("配置文件中未定义 repos")
			return nil
		}

		ui.Title("Git 仓库")
		fmt.Println()

		syncRepos(cfg.Repos, dryRun)

		fmt.Println()
		ui.Success("仓库同步完成")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
}
