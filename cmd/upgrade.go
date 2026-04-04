package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/chzealot/kickstart/internal/cache"
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
		result := checker.WaitResult(10 * time.Second)
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
		// 升级成功后跳过 PersistentPostRun 的版本提示
		versionChecker = nil
		return nil
	},
}

func doUpgrade(tag string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	asset := fmt.Sprintf("kickstart_%s_%s", goos, goarch)
	if goos == "windows" {
		asset += ".exe"
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 1. Check local cache
	if cachedPath, ok := cache.HasValidBinary(tag, asset); ok {
		ui.Info("使用本地缓存: %s", cachedPath)
		if err := installBinary(cachedPath, execPath); err != nil {
			return err
		}
		cleanOldCache()
		return nil
	}

	// 2. Download checksums.txt and binary
	if err := downloadAndCache(tag, asset); err != nil {
		return err
	}

	// 3. Verify checksum
	if err := cache.VerifyChecksum(tag, asset); err != nil {
		return fmt.Errorf("checksum 校验失败: %w", err)
	}

	// 4. Install from cache
	cacheDir, err := cache.Dir(tag)
	if err != nil {
		return err
	}
	cachedBinary := cacheDir + "/" + asset
	if err := installBinary(cachedBinary, execPath); err != nil {
		return err
	}

	// 5. Clean old versions
	cleanOldCache()
	return nil
}

func downloadAndCache(tag, asset string) error {
	if _, err := exec.LookPath("gh"); err == nil {
		return downloadWithGH(tag, asset)
	}
	return downloadWithCurl(tag, asset)
}

func downloadWithGH(tag, asset string) error {
	cacheDir, err := cache.Dir(tag)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// Download checksum file
	cmd := exec.Command("gh", "release", "download", tag,
		"--repo", "chzealot/kickstart",
		"-p", "checksums.txt",
		"-D", cacheDir,
		"--clobber")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载 checksums.txt 失败: %w", err)
	}

	// Download binary
	cmd = exec.Command("gh", "release", "download", tag,
		"--repo", "chzealot/kickstart",
		"-p", asset,
		"-D", cacheDir,
		"--clobber")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载二进制文件失败: %w", err)
	}

	return nil
}

func downloadWithCurl(tag, asset string) error {
	cacheDir, err := cache.Dir(tag)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	baseURL := fmt.Sprintf("https://github.com/chzealot/kickstart/releases/download/%s", tag)

	// Download checksums.txt
	if err := curlDownload(baseURL+"/checksums.txt", cacheDir+"/checksums.txt"); err != nil {
		return fmt.Errorf("下载 checksums.txt 失败: %w", err)
	}

	// Download binary
	if err := curlDownload(baseURL+"/"+asset, cacheDir+"/"+asset); err != nil {
		return fmt.Errorf("下载二进制文件失败: %w", err)
	}

	return nil
}

func curlDownload(url, dest string) error {
	args := []string{"-fsSL", "-o", dest}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		args = append(args, "-H", "Authorization: token "+token)
	}
	args = append(args, url)

	cmd := exec.Command("curl", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installBinary(src, dest string) error {
	if err := os.Chmod(src, 0755); err != nil {
		return err
	}

	// 覆盖前校验：运行新二进制的 --version 确认可执行
	if out, err := exec.Command(src, "--version").CombinedOutput(); err != nil {
		return fmt.Errorf("下载的二进制无法正常运行: %s (%w)", strings.TrimSpace(string(out)), err)
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

func cleanOldCache() {
	if err := cache.CleanOldVersions(3); err != nil {
		ui.Warn("清理历史版本缓存失败: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
