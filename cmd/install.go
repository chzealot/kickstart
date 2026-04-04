package cmd

import (
	"fmt"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/installer"
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
			ui.Warn("配置文件不存在: %s", cfg.Path)
			ui.Dim("请创建配置文件，示例：")
			ui.Dim("  tools:")
			ui.Dim("    - rsync")
			ui.Dim("    - jq")
			return nil
		}

		if len(cfg.Tools) == 0 {
			ui.Info("配置文件中未定义 tools")
			return nil
		}

		ui.Title("安装工具和软件包")
		fmt.Println()

		tools := installer.FromNames(cfg.Tools)
		hasError := false
		for _, tool := range tools {
			if tool.Check() {
				ui.Success("%s 已安装", tool.Name)
				continue
			}

			if dryRun {
				ui.Step("将安装 %s（dry-run 模式，跳过）", tool.Name)
				continue
			}

			sp := ui.StartSpinner(fmt.Sprintf("正在安装 %s...", tool.Name))
			err := tool.Install(false)
			sp.Stop()

			if err != nil {
				ui.Error("安装 %s 失败: %v", tool.Name, err)
				hasError = true
			} else {
				ui.Success("%s 安装成功", tool.Name)
			}
		}

		fmt.Println()
		if hasError {
			return fmt.Errorf("部分工具安装失败")
		}
		ui.Success("所有工具已就绪")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
