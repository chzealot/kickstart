package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chzealot/kickstart/internal/config"
	"github.com/chzealot/kickstart/internal/ui"
)

// promptInitConfig prompts the user to create a default config file.
// Returns true if the config was created, false otherwise.
func promptInitConfig(cfg *config.Config) bool {
	ui.Warn("配置文件不存在: %s", cfg.Path)
	fmt.Println()
	fmt.Printf("是否初始化配置文件? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		ui.Dim("跳过初始化，参考 README 手动创建配置文件")
		return false
	}

	if err := config.Init(cfg.Path); err != nil {
		ui.Error("初始化配置失败: %v", err)
		return false
	}

	ui.Success("已创建配置文件: %s", cfg.Path)
	ui.Dim("请编辑配置文件后重新运行 kickstart")
	return true
}
