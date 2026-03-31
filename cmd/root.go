package cmd

import (
	"fmt"
	"os"

	"github.com/chzealot/kickstart/internal/ui"
	"github.com/chzealot/kickstart/internal/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dryRun  bool
	verbose bool

	versionChecker *version.AsyncChecker
)

var rootCmd = &cobra.Command{
	Use:   "kickstart",
	Short: "一键初始化新电脑环境",
	Long:  "kickstart - 一键初始化新电脑环境的命令行工具",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Start background version check
		versionChecker = version.NewAsyncChecker()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Show upgrade hint if a new version is available
		if versionChecker == nil {
			return
		}
		if result := versionChecker.Result(); result != nil && result.HasUpdate {
			fmt.Println()
			ui.Warn("kickstart %s 已发布（当前版本 %s），运行 kickstart upgrade 升级",
				result.Latest, result.Current)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径（默认 ~/.kickstart.yaml）")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "仅预览变更，不实际执行")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "输出详细日志")

	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate("kickstart version {{.Version}}\n")
}
