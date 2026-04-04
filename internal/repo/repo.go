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

// SyncDotfiles manages dotfiles as a bare repo in ~/.git.
// The work tree is $HOME, so dotfiles are checked out directly into ~.
// If ~/.git doesn't exist, clone as bare; otherwise fetch and merge.
func SyncDotfiles(url string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %w", err)
	}

	gitDir := filepath.Join(home, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return cloneDotfiles(url, home, gitDir)
	}

	return pullDotfiles(home, gitDir)
}

func cloneDotfiles(url, home, gitDir string) error {
	// Clone as bare repo
	cmd := exec.Command("git", "clone", "--bare", url, gitDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("clone 失败: %s", strings.TrimSpace(string(out)))
	}

	// Checkout files to home directory
	cmd = exec.Command("git", "--git-dir="+gitDir, "--work-tree="+home, "checkout")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Checkout may fail if files already exist; try with force
		cmd = exec.Command("git", "--git-dir="+gitDir, "--work-tree="+home, "checkout", "-f")
		if out2, err2 := cmd.CombinedOutput(); err2 != nil {
			return fmt.Errorf("checkout 失败: %s", strings.TrimSpace(string(append(out, out2...))))
		}
	}

	// Configure bare repo to not show untracked files
	cmd = exec.Command("git", "--git-dir="+gitDir, "--work-tree="+home,
		"config", "status.showUntrackedFiles", "no")
	cmd.CombinedOutput()

	return nil
}

func pullDotfiles(home, gitDir string) error {
	cmd := exec.Command("git", "--git-dir="+gitDir, "--work-tree="+home, "pull", "--ff-only")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pull 失败: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func clone(url, path string) error {
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	cmd := exec.Command("git", "clone", url, path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("clone 失败: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func pull(path string) error {
	cmd := exec.Command("git", "-C", path, "pull", "--ff-only")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pull 失败: %s", strings.TrimSpace(string(out)))
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
