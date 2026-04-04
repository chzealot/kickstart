package cache

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	cacheSubdir      = "kickstart"
	checksumFileName = "checksums.txt"
	maxKeepVersions  = 3
)

// Dir returns the cache directory for a given version: ~/.cache/kickstart/{version}/
func Dir(version string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(home, ".cache", cacheSubdir, version), nil
}

// HasValidBinary checks if a cached binary exists and its checksum matches.
func HasValidBinary(version, asset string) (string, bool) {
	dir, err := Dir(version)
	if err != nil {
		return "", false
	}

	binaryPath := filepath.Join(dir, asset)
	if _, err := os.Stat(binaryPath); err != nil {
		return "", false
	}

	if err := VerifyChecksum(version, asset); err != nil {
		return "", false
	}

	return binaryPath, true
}

// SaveFile writes data to the cache directory for a given version.
func SaveFile(version, filename string, data []byte) error {
	dir, err := Dir(version)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	path := filepath.Join(dir, filename)
	return os.WriteFile(path, data, 0644)
}

// SaveChecksum writes the checksums.txt to the cache directory.
func SaveChecksum(version string, data []byte) error {
	return SaveFile(version, checksumFileName, data)
}

// VerifyChecksum reads checksums.txt from cache and verifies the binary's sha256.
func VerifyChecksum(version, asset string) error {
	dir, err := Dir(version)
	if err != nil {
		return err
	}

	expected, err := readExpectedChecksum(filepath.Join(dir, checksumFileName), asset)
	if err != nil {
		return err
	}

	actual, err := hashFile(filepath.Join(dir, asset))
	if err != nil {
		return fmt.Errorf("计算文件哈希失败: %w", err)
	}

	if actual != expected {
		return fmt.Errorf("校验失败: 期望 %s, 实际 %s", expected, actual)
	}
	return nil
}

// CleanOldVersions keeps only the most recent `keep` versions by modification time.
func CleanOldVersions(keep int) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	baseDir := filepath.Join(home, ".cache", cacheSubdir)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	type versionEntry struct {
		name    string
		modTime int64
	}

	var versions []versionEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		versions = append(versions, versionEntry{name: e.Name(), modTime: info.ModTime().UnixNano()})
	}

	if len(versions) <= keep {
		return nil
	}

	// Sort by modification time descending (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].modTime > versions[j].modTime
	})

	// Remove old versions
	var errs []string
	for _, v := range versions[keep:] {
		if err := os.RemoveAll(filepath.Join(baseDir, v.name)); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", v.name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分版本清理失败: %s", strings.Join(errs, "; "))
	}
	return nil
}

// readExpectedChecksum parses checksums.txt and returns the expected hash for asset.
// Format: "<sha256>  <filename>" (two spaces, matching sha256sum output)
func readExpectedChecksum(checksumPath, asset string) (string, error) {
	f, err := os.Open(checksumPath)
	if err != nil {
		return "", fmt.Errorf("读取 checksum 文件失败: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: hash  filename (two spaces) or hash *filename (binary mode)
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimPrefix(parts[1], "*")
		if name == asset {
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("checksum 文件中未找到 %s", asset)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
