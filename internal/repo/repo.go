package repo

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Sync clones the repo if the target path does not exist,
// or pulls updates if it already exists.
func Sync(url, path string) error {
	path = ExpandHome(path)

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
		// Backup conflicting files before force checkout
		backupDir, backupErr := backupConflictingFiles(home, string(out))
		if backupErr != nil {
			return fmt.Errorf("备份冲突文件失败: %w\ncheckout 原始错误: %s", backupErr, strings.TrimSpace(string(out)))
		}

		cmd = exec.Command("git", "--git-dir="+gitDir, "--work-tree="+home, "checkout", "-f")
		if out2, err2 := cmd.CombinedOutput(); err2 != nil {
			return fmt.Errorf("checkout 失败: %s", strings.TrimSpace(string(append(out, out2...))))
		}

		if backupDir != "" {
			fmt.Fprintf(os.Stderr, "  已备份冲突文件到: %s\n", backupDir)
		}
	}

	// Configure bare repo to not show untracked files
	cmd = exec.Command("git", "--git-dir="+gitDir, "--work-tree="+home,
		"config", "status.showUntrackedFiles", "no")
	if configOut, configErr := cmd.CombinedOutput(); configErr != nil {
		fmt.Fprintf(os.Stderr, "  警告: 设置 status.showUntrackedFiles 失败: %s\n", strings.TrimSpace(string(configOut)))
	}

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

// backupConflictingFiles parses git checkout error output to find conflicting files,
// backs them up to ~/.kickstart-backup/<timestamp>/, and returns the backup directory.
func backupConflictingFiles(home, checkoutOutput string) (string, error) {
	files := parseConflictingFiles(checkoutOutput)
	if len(files) == 0 {
		return "", nil
	}

	timestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(home, ".kickstart-backup", timestamp)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("创建备份目录失败: %w", err)
	}

	for _, f := range files {
		src := filepath.Join(home, f)
		dst := filepath.Join(backupDir, f)

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			continue
		}

		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		os.WriteFile(dst, data, 0644)
	}

	return backupDir, nil
}

// parseConflictingFiles extracts file paths from git checkout error output.
// Git outputs lines like: "\terror: ... would be overwritten by checkout:\n\t\t.bashrc\n"
func parseConflictingFiles(output string) []string {
	var files []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	inFileList := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "would be overwritten by checkout") {
			inFileList = true
			continue
		}
		if inFileList {
			if line == "" || strings.HasPrefix(line, "Please") || strings.HasPrefix(line, "error:") || strings.HasPrefix(line, "Aborting") {
				inFileList = false
				continue
			}
			files = append(files, line)
		}
	}
	return files
}

// ExpandHome expands ~ prefix in a path to the user's home directory.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
