package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// archiveName returns the archive filename for the current platform.
// e.g. kickstart-darwin-arm64.tar.gz or kickstart-windows-amd64.zip
func archiveName() string {
	name := fmt.Sprintf("kickstart-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		return name + ".zip"
	}
	return name + ".tar.gz"
}

// binaryName returns the binary filename inside the archive.
func binaryName() string {
	if runtime.GOOS == "windows" {
		return "kickstart.exe"
	}
	return "kickstart"
}

// dirName returns the directory name inside the archive.
func dirName() string {
	return fmt.Sprintf("kickstart-%s-%s", runtime.GOOS, runtime.GOARCH)
}

func doUpgrade(tag string) error {
	archive := archiveName()
	binary := binaryName()

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 1. Check local cache for extracted binary
	if cachedPath, ok := cache.HasValidBinary(tag, archive); ok {
		ui.Info("使用本地缓存")
		binaryPath := filepath.Join(filepath.Dir(cachedPath), dirName(), binary)
		if _, err := os.Stat(binaryPath); err == nil {
			return installBinary(binaryPath, execPath)
		}
		// Fallback: extract again
	}

	// 2. Download checksums.txt and archive
	if err := downloadAndCache(tag, archive); err != nil {
		return err
	}

	// 3. Verify checksum
	if err := cache.VerifyChecksum(tag, archive); err != nil {
		return fmt.Errorf("checksum 校验失败: %w", err)
	}

	// 4. Extract binary from archive
	cacheDir, err := cache.Dir(tag)
	if err != nil {
		return err
	}
	archivePath := filepath.Join(cacheDir, archive)
	if err := extractArchive(archivePath, cacheDir); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 5. Install binary
	binaryPath := filepath.Join(cacheDir, dirName(), binary)
	if err := installBinary(binaryPath, execPath); err != nil {
		return err
	}

	// 6. Clean old versions
	cleanOldCache()
	return nil
}

func extractArchive(archivePath, destDir string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		cmd := exec.Command("unzip", "-o", archivePath, "-d", destDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}

	// tar.gz
	cmd := exec.Command("tar", "xzf", archivePath, "-C", destDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func downloadAndCache(tag, archive string) error {
	if _, err := exec.LookPath("gh"); err == nil {
		return downloadWithGH(tag, archive)
	}
	return downloadWithCurl(tag, archive)
}

func downloadWithGH(tag, archive string) error {
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

	// Download archive
	cmd = exec.Command("gh", "release", "download", tag,
		"--repo", "chzealot/kickstart",
		"-p", archive,
		"-D", cacheDir,
		"--clobber")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载压缩包失败: %w", err)
	}

	return nil
}

func downloadWithCurl(tag, archive string) error {
	cacheDir, err := cache.Dir(tag)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	baseURL := fmt.Sprintf("https://github.com/chzealot/kickstart/releases/download/%s", tag)

	// Download checksums.txt
	if err := curlDownload(baseURL+"/checksums.txt", filepath.Join(cacheDir, "checksums.txt")); err != nil {
		return fmt.Errorf("下载 checksums.txt 失败: %w", err)
	}

	// Download archive
	if err := curlDownload(baseURL+"/"+archive, filepath.Join(cacheDir, archive)); err != nil {
		return fmt.Errorf("下载压缩包失败: %w", err)
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
