package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/repo"
	"github.com/chzealot/kickstart/internal/ui"
	"github.com/spf13/cobra"
)

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "克隆或更新 Git 仓库",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			ui.Error("加载配置失败: %v", err)
			return nil
		}

		if !cfg.Exists() {
			ui.Warn("配置文件不存在: %s", cfg.Path)
			ui.Dim("请创建 ~/.kickstart/config.yaml，参考 README 中的配置说明")
			return nil
		}

		if len(cfg.Repos) == 0 {
			ui.Info("配置文件中未定义 repos")
			return nil
		}

		ui.Title("克隆或更新 Git 仓库")
		fmt.Println()

		hasError := false
		for _, r := range cfg.Repos {
			if dryRun {
				ui.Step("将同步 %s → %s（dry-run 模式，跳过）", r.URL, r.Path)
				continue
			}

			sp := ui.StartSpinner(fmt.Sprintf("同步 %s ...", r.URL))
			err := repo.Sync(r.URL, r.Path)
			sp.Stop()

			if err != nil {
				ui.Error("%s → %s 失败: %v", r.URL, r.Path, err)
				hasError = true
			} else {
				ui.Success("%s → %s", r.URL, r.Path)
			}
		}

		fmt.Println()
		if hasError {
			return fmt.Errorf("部分仓库同步失败")
		}
		ui.Success("所有仓库已同步")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
}
