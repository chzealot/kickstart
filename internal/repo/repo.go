package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Sync clones the repo if the target path does not exist,
// or pulls updates if it already exists.
func Sync(url, path string) error {
	path = expandHome(path)

	if isGitRepo(path) {
		return pull(path)
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("目录已存在但不是 Git 仓库: %s", path)
	}

	return clone(url, path)
}

func clone(url, path string) error {
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	cmd := exec.Command("git", "clone", url, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clone 失败: %w", err)
	}
	return nil
}

func pull(path string) error {
	cmd := exec.Command("git", "-C", path, "pull", "--ff-only")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pull 失败: %w", err)
	}
	return nil
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	// .git can be a directory (regular repo) or a file (worktree/submodule)
	return info.IsDir() || info.Mode().IsRegular()
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
