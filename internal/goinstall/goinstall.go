package goinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Sources for Go downloads, in priority order.
var sources = []string{
	"https://go.dev/dl/",
	"https://golang.google.cn/dl/",
}

// apiEndpoints for fetching version info, in priority order.
var apiEndpoints = []string{
	"https://go.dev/dl/?mode=json",
	"https://golang.google.cn/dl/?mode=json",
}

// installDir is where Go gets installed.
const installDir = "/usr/local/go"

// Release represents a Go release from the download API.
type Release struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []File `json:"files"`
}

// File represents a downloadable file in a release.
type File struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

// FetchLatestVersion fetches the latest stable Go version and its file list.
// Tries go.dev first, falls back to golang.google.cn.
func FetchLatestVersion() (string, []File, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	for _, endpoint := range apiEndpoints {
		releases, err := fetchReleases(client, endpoint)
		if err != nil {
			continue
		}
		for _, r := range releases {
			if r.Stable {
				return r.Version, r.Files, nil
			}
		}
	}

	return "", nil, fmt.Errorf("无法获取 Go 最新版本（go.dev 和 golang.google.cn 均不可用）")
}

func fetchReleases(client *http.Client, url string) ([]Release, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

// LocalVersion returns the installed Go version (e.g. "go1.26.1") or "" if not installed.
func LocalVersion() string {
	goBin := filepath.Join(installDir, "bin", "go")
	out, err := exec.Command(goBin, "version").Output()
	if err != nil {
		return ""
	}
	// Output format: "go version go1.26.1 linux/amd64"
	parts := strings.Fields(string(out))
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// NeedsInstall returns true if Go needs to be installed or updated.
func NeedsInstall(latest, local string) bool {
	return local == "" || local != latest
}

// FindArchive finds the matching archive file for the current platform.
func FindArchive(files []File) (*File, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	for i := range files {
		f := &files[i]
		if f.Kind == "archive" && f.OS == goos && f.Arch == goarch {
			return f, nil
		}
	}
	return nil, fmt.Errorf("未找到适用于 %s/%s 的 Go 安装包", goos, goarch)
}

// Install downloads and installs the specified Go version.
func Install(file *File) error {
	// 1. Download to temp file
	tmpFile, err := downloadArchive(file.Filename)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer os.Remove(tmpFile)

	// 2. Verify SHA256
	if err := verifyChecksum(tmpFile, file.SHA256); err != nil {
		return fmt.Errorf("校验失败: %w", err)
	}

	// 3. Remove old installation
	if _, err := os.Stat(installDir); err == nil {
		if err := runSudo("rm", "-rf", installDir); err != nil {
			return fmt.Errorf("删除旧版本失败: %w", err)
		}
	}

	// 4. Extract to /usr/local
	if err := runSudo("tar", "-C", "/usr/local", "-xzf", tmpFile); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 5. Verify installation
	goBin := filepath.Join(installDir, "bin", "go")
	out, err := exec.Command(goBin, "version").Output()
	if err != nil {
		return fmt.Errorf("安装验证失败: %w", err)
	}
	_ = out

	return nil
}

// downloadArchive downloads the Go archive, trying sources in order.
// Returns the path to the downloaded temp file.
func downloadArchive(filename string) (string, error) {
	tmpFile, err := os.CreateTemp("", "kickstart-go-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	client := &http.Client{Timeout: 10 * time.Minute}

	for _, base := range sources {
		url := base + filename
		if err := downloadToFile(client, url, tmpPath); err != nil {
			continue
		}
		return tmpPath, nil
	}

	os.Remove(tmpPath)
	return "", fmt.Errorf("所有下载源均失败（go.dev 和 golang.google.cn）")
}

func downloadToFile(client *http.Client, url, dest string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func verifyChecksum(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("SHA256 不匹配\n  期望: %s\n  实际: %s", expected, actual)
	}
	return nil
}

func runSudo(args ...string) error {
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
