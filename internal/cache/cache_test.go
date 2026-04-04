package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func sha256sum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func TestVerifyChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	versionDir := filepath.Join(tmpDir, "v1.0.0")
	os.MkdirAll(versionDir, 0755)

	binaryData := []byte("fake binary content")
	binaryHash := sha256sum(binaryData)
	asset := "kickstart_darwin_arm64"

	os.WriteFile(filepath.Join(versionDir, asset), binaryData, 0644)

	t.Run("valid checksum", func(t *testing.T) {
		checksumContent := fmt.Sprintf("%s  %s\n", binaryHash, asset)
		os.WriteFile(filepath.Join(versionDir, checksumFileName), []byte(checksumContent), 0644)

		err := verifyChecksumInDir(versionDir, asset)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid checksum", func(t *testing.T) {
		checksumContent := fmt.Sprintf("%s  %s\n", "0000000000000000000000000000000000000000000000000000000000000000", asset)
		os.WriteFile(filepath.Join(versionDir, checksumFileName), []byte(checksumContent), 0644)

		err := verifyChecksumInDir(versionDir, asset)
		if err == nil {
			t.Fatal("expected error for mismatched checksum")
		}
	})

	t.Run("missing asset in checksum file", func(t *testing.T) {
		checksumContent := fmt.Sprintf("%s  %s\n", binaryHash, "other_asset")
		os.WriteFile(filepath.Join(versionDir, checksumFileName), []byte(checksumContent), 0644)

		err := verifyChecksumInDir(versionDir, asset)
		if err == nil {
			t.Fatal("expected error for missing asset")
		}
	})

	t.Run("missing checksum file", func(t *testing.T) {
		os.Remove(filepath.Join(versionDir, checksumFileName))

		err := verifyChecksumInDir(versionDir, asset)
		if err == nil {
			t.Fatal("expected error for missing checksum file")
		}
	})
}

// verifyChecksumInDir is a test helper that verifies checksum in a specific directory.
func verifyChecksumInDir(dir, asset string) error {
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

func TestCleanOldVersions(t *testing.T) {
	// Override home dir via env
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	baseDir := filepath.Join(tmpHome, ".cache", cacheSubdir)
	os.MkdirAll(baseDir, 0755)

	// Create 5 version directories with different mod times
	versions := []string{"v0.1.0", "v0.2.0", "v0.3.0", "v0.4.0", "v0.5.0"}
	for i, v := range versions {
		dir := filepath.Join(baseDir, v)
		os.MkdirAll(dir, 0755)
		// Write a marker file so we can verify deletion
		os.WriteFile(filepath.Join(dir, "marker"), []byte(v), 0644)
		// Set mod time: older versions get earlier times
		modTime := time.Now().Add(time.Duration(i) * time.Minute)
		os.Chtimes(dir, modTime, modTime)
	}

	err := CleanOldVersions(3)
	if err != nil {
		t.Fatalf("CleanOldVersions failed: %v", err)
	}

	// Newest 3 should remain: v0.3.0, v0.4.0, v0.5.0
	remaining, _ := os.ReadDir(baseDir)
	if len(remaining) != 3 {
		names := make([]string, len(remaining))
		for i, e := range remaining {
			names[i] = e.Name()
		}
		t.Fatalf("expected 3 versions remaining, got %d: %v", len(remaining), names)
	}

	for _, v := range []string{"v0.3.0", "v0.4.0", "v0.5.0"} {
		if _, err := os.Stat(filepath.Join(baseDir, v)); err != nil {
			t.Errorf("expected %s to remain, but it was deleted", v)
		}
	}

	for _, v := range []string{"v0.1.0", "v0.2.0"} {
		if _, err := os.Stat(filepath.Join(baseDir, v)); err == nil {
			t.Errorf("expected %s to be deleted, but it still exists", v)
		}
	}
}

func TestCleanOldVersions_FewVersions(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	baseDir := filepath.Join(tmpHome, ".cache", cacheSubdir)
	os.MkdirAll(baseDir, 0755)

	// Only 2 versions - nothing should be deleted
	for _, v := range []string{"v0.1.0", "v0.2.0"} {
		os.MkdirAll(filepath.Join(baseDir, v), 0755)
	}

	err := CleanOldVersions(3)
	if err != nil {
		t.Fatalf("CleanOldVersions failed: %v", err)
	}

	remaining, _ := os.ReadDir(baseDir)
	if len(remaining) != 2 {
		t.Fatalf("expected 2 versions remaining, got %d", len(remaining))
	}
}

func TestCleanOldVersions_EmptyDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// No cache dir at all - should not error
	err := CleanOldVersions(3)
	if err != nil {
		t.Fatalf("CleanOldVersions failed: %v", err)
	}
}

func TestReadExpectedChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	checksumFile := filepath.Join(tmpDir, "checksums.txt")

	content := `abc123def456  kickstart_darwin_amd64
789xyz000111  kickstart_darwin_arm64
fedcba987654  kickstart_linux_amd64
`
	os.WriteFile(checksumFile, []byte(content), 0644)

	tests := []struct {
		asset    string
		expected string
		wantErr  bool
	}{
		{"kickstart_darwin_amd64", "abc123def456", false},
		{"kickstart_darwin_arm64", "789xyz000111", false},
		{"kickstart_linux_amd64", "fedcba987654", false},
		{"kickstart_windows_amd64", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.asset, func(t *testing.T) {
			got, err := readExpectedChecksum(checksumFile, tt.asset)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}
