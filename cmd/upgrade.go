package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/chzealot/kickstart/internal/ui"
	"github.com/chzealot/kickstart/internal/version"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "更新 kickstart 工具自身到最新版本",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("升级 kickstart")
		fmt.Println()
		ui.Info("当前版本: %s", version.Version)

		sp := ui.StartSpinner("检查最新版本...")
		checker := version.NewAsyncChecker()
		// Wait a bit for the result
		var result *version.CheckResult
		for i := 0; i < 100; i++ {
			result = checker.Result()
			if result != nil {
				break
			}
			// small busy wait
			runtime.Gosched()
		}
		sp.Stop()

		if result == nil {
			ui.Error("无法获取最新版本信息")
			return nil
		}

		if !result.HasUpdate {
			ui.Success("已是最新版本")
			return nil
		}

		ui.Info("最新版本: %s", result.Latest)
		fmt.Println()
		ui.Step("正在下载 kickstart %s ...", result.Latest)

		if err := doUpgrade(result.Latest); err != nil {
			ui.Error("升级失败: %v", err)
			return nil
		}

		ui.Success("升级成功！当前版本: %s", result.Latest)
		return nil
	},
}

func doUpgrade(tag string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	asset := fmt.Sprintf("kickstart_%s_%s", goos, goarch)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// Use GitHub CLI if available, otherwise fall back to curl
	if _, err := exec.LookPath("gh"); err == nil {
		return upgradeWithGH(asset, execPath, tag)
	}
	return upgradeWithCurl(asset, execPath, tag)
}

func upgradeWithGH(asset, dest, tag string) error {
	tmpDir, err := os.MkdirTemp("", "kickstart-upgrade-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("gh", "release", "download", tag,
		"--repo", "chzealot/kickstart",
		"-p", asset,
		"-D", tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	return installBinary(tmpDir+"/"+asset, dest)
}

func upgradeWithCurl(asset, dest, tag string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("需要设置 GITHUB_TOKEN 环境变量（或安装 gh CLI）")
	}

	url := fmt.Sprintf("https://github.com/chzealot/kickstart/releases/download/%s/%s", tag, asset)

	tmpFile, err := os.CreateTemp("", "kickstart-*")
	if err != nil {
		return err
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	cmd := exec.Command("curl", "-fsSL",
		"-H", "Authorization: token "+token,
		"-H", "Accept: application/octet-stream",
		"-o", tmpFile.Name(),
		url)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	return installBinary(tmpFile.Name(), dest)
}

func installBinary(src, dest string) error {
	if err := os.Chmod(src, 0755); err != nil {
		return err
	}

	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dest, input, 0755); err != nil {
		return fmt.Errorf("写入 %s 失败（可能需要 sudo）: %w", dest, err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
