package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const (
	defaultConfigDir  = ".kickstart"
	defaultConfigFile = "config.yaml"
)

// Section holds configuration for a scope (global, platform, or host).
type Section struct {
	Dotfiles *DotfilesConfig `yaml:"dotfiles,omitempty"`
	Go       string          `yaml:"go,omitempty"`
	Repos    []RepoConfig    `yaml:"repos,omitempty"`
	Tools    []string        `yaml:"tools,omitempty"`
	Scripts  []ScriptTask    `yaml:"scripts,omitempty"`
}

// rawConfig is the full YAML structure before merging.
type rawConfig struct {
	Include []string `yaml:"include,omitempty"`
	Section `yaml:",inline"`
	Darwin  *Section            `yaml:"darwin,omitempty"`
	Linux   *Section            `yaml:"linux,omitempty"`
	Windows *Section            `yaml:"windows,omitempty"`
	Hosts   map[string]*Section `yaml:"hosts,omitempty"`
}

// Config holds the merged configuration ready for use.
type Config struct {
	Path     string
	Dotfiles *DotfilesConfig
	Go       string
	Repos    []RepoConfig
	Tools    []string
	Scripts  []ScriptTask
}

// DotfilesConfig holds dotfiles repository settings.
type DotfilesConfig struct {
	Repo string `yaml:"repo"`
}

// RepoConfig holds a git repository to clone/update.
type RepoConfig struct {
	URL  string `yaml:"url"`
	Path string `yaml:"path"`
}

// ScriptTask holds a named shell command to run.
type ScriptTask struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

// Load reads, parses, and merges the config file.
// Merge order: includes → global → platform (darwin/linux/windows) → matching hosts.
// Path resolution:
//   - empty: defaults to ~/.kickstart/config.yaml (falls back to ~/.kickstart if it's a file)
//   - directory: loads config.yaml inside it
//   - file: loads that file directly
func Load(path string) (*Config, error) {
	return loadWithEnv(path, runtime.GOOS, hostname())
}

// loadWithEnv is the testable core of Load.
func loadWithEnv(path, goos, host string) (*Config, error) {
	// Clear warnings from previous loads
	duplicateWarnings = nil

	path, err := resolveConfigPath(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{Path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	raw, err := loadRawWithIncludes(path, data, make(map[string]bool))
	if err != nil {
		return nil, err
	}

	// Start with global section
	merged := &Section{
		Dotfiles: raw.Section.Dotfiles,
		Go:       raw.Section.Go,
		Repos:    append([]RepoConfig{}, raw.Section.Repos...),
		Tools:    append([]string{}, raw.Section.Tools...),
		Scripts:  append([]ScriptTask{}, raw.Section.Scripts...),
	}

	// Merge platform section
	var platformSection *Section
	switch goos {
	case "darwin":
		platformSection = raw.Darwin
	case "linux":
		platformSection = raw.Linux
	case "windows":
		platformSection = raw.Windows
	}
	if platformSection != nil {
		mergeSection(merged, platformSection)
	}

	// Merge matching host sections
	if host != "" && len(raw.Hosts) > 0 {
		for pattern, section := range raw.Hosts {
			if section == nil {
				continue
			}
			matched, err := filepath.Match(pattern, host)
			if err != nil {
				continue // invalid pattern, skip
			}
			if matched {
				mergeSection(merged, section)
			}
		}
	}

	cfg.Dotfiles = merged.Dotfiles
	cfg.Go = merged.Go
	cfg.Repos = merged.Repos
	cfg.Tools = merged.Tools
	cfg.Scripts = merged.Scripts
	return cfg, nil
}

// resolveConfigPath determines the actual config file path.
func resolveConfigPath(path string) (string, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("获取用户主目录失败: %w", err)
		}
		dirPath := filepath.Join(home, defaultConfigDir)
		filePath := filepath.Join(dirPath, defaultConfigFile)

		// Check if ~/.kickstart/ directory exists
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			return filePath, nil
		}

		// Fall back to ~/.kickstart as a file (legacy)
		if info, err := os.Stat(dirPath); err == nil && !info.IsDir() {
			return dirPath, nil
		}

		// Neither exists, default to directory style
		return filePath, nil
	}

	// If path is a directory, load config.yaml inside it
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return filepath.Join(path, defaultConfigFile), nil
	}

	return path, nil
}

// loadRawWithIncludes loads a rawConfig and recursively processes includes.
// visited tracks absolute paths to detect cycles.
func loadRawWithIncludes(path string, data []byte, visited map[string]bool) (*rawConfig, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("解析路径失败: %w", err)
	}

	if visited[absPath] {
		return nil, fmt.Errorf("检测到循环引用: %s", path)
	}
	visited[absPath] = true

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析配置文件失败 (%s): %w", path, err)
	}

	if len(raw.Include) == 0 {
		return &raw, nil
	}

	baseDir := filepath.Dir(absPath)

	// Process includes: each included file is loaded and merged into raw
	for _, inc := range raw.Include {
		incPath := inc
		if !filepath.IsAbs(incPath) {
			incPath = filepath.Join(baseDir, incPath)
		}

		incData, err := os.ReadFile(incPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // skip missing include files
			}
			return nil, fmt.Errorf("读取 include 文件失败 (%s): %w", inc, err)
		}

		incRaw, err := loadRawWithIncludes(incPath, incData, visited)
		if err != nil {
			return nil, err
		}

		mergeRawConfig(&raw, incRaw)
	}

	return &raw, nil
}

// mergeRawConfig merges src rawConfig into dst.
func mergeRawConfig(dst, src *rawConfig) {
	// Merge global section
	mergeSection(&dst.Section, &src.Section)

	// Merge platform sections
	dst.Darwin = mergeSectionPtr(dst.Darwin, src.Darwin)
	dst.Linux = mergeSectionPtr(dst.Linux, src.Linux)
	dst.Windows = mergeSectionPtr(dst.Windows, src.Windows)

	// Merge hosts
	if len(src.Hosts) > 0 {
		if dst.Hosts == nil {
			dst.Hosts = make(map[string]*Section)
		}
		for pattern, section := range src.Hosts {
			if section == nil {
				continue
			}
			dst.Hosts[pattern] = mergeSectionPtr(dst.Hosts[pattern], section)
		}
	}
}

// mergeSectionPtr merges two *Section, handling nil cases.
func mergeSectionPtr(dst, src *Section) *Section {
	if src == nil {
		return dst
	}
	if dst == nil {
		cp := *src
		return &cp
	}
	mergeSection(dst, src)
	return dst
}

// mergeSection merges src into dst.
// Tools, Repos, Scripts are appended (with deduplication). Dotfiles and Go are overridden if set.
func mergeSection(dst, src *Section) {
	if src.Dotfiles != nil {
		dst.Dotfiles = src.Dotfiles
	}
	if src.Go != "" {
		dst.Go = src.Go
	}
	dst.Repos = mergeRepos(dst.Repos, src.Repos)
	dst.Tools = mergeTools(dst.Tools, src.Tools)
	dst.Scripts = append(dst.Scripts, src.Scripts...)
}

// mergeTools appends new tools, skipping duplicates and recording warnings.
func mergeTools(dst, src []string) []string {
	seen := make(map[string]bool, len(dst))
	for _, t := range dst {
		seen[t] = true
	}
	for _, t := range src {
		if seen[t] {
			addDuplicateWarning(fmt.Sprintf("tools 中存在重复项: %s（已自动去重）", t))
			continue
		}
		seen[t] = true
		dst = append(dst, t)
	}
	return dst
}

// mergeRepos appends new repos, skipping duplicates by path and recording warnings.
func mergeRepos(dst, src []RepoConfig) []RepoConfig {
	seen := make(map[string]bool, len(dst))
	for _, r := range dst {
		seen[r.Path] = true
	}
	for _, r := range src {
		if seen[r.Path] {
			addDuplicateWarning(fmt.Sprintf("repos 中存在重复路径: %s（已自动去重）", r.Path))
			continue
		}
		seen[r.Path] = true
		dst = append(dst, r)
	}
	return dst
}

// DuplicateWarnings stores warnings generated during config merging.
var duplicateWarnings []string

func addDuplicateWarning(msg string) {
	// Avoid duplicating the warning itself
	for _, w := range duplicateWarnings {
		if w == msg {
			return
		}
	}
	duplicateWarnings = append(duplicateWarnings, msg)
}

// PopDuplicateWarnings returns and clears accumulated duplicate warnings.
func PopDuplicateWarnings() []string {
	w := duplicateWarnings
	duplicateWarnings = nil
	return w
}

func hostname() string {
	name, _ := os.Hostname()
	return name
}

// DefaultPath returns the default config file path.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/" + defaultConfigDir + "/" + defaultConfigFile
	}
	return filepath.Join(home, defaultConfigDir, defaultConfigFile)
}

// Exists reports whether the config file exists.
func (c *Config) Exists() bool {
	_, err := os.Stat(c.Path)
	return err == nil
}

const defaultConfigTemplate = `# kickstart configuration
# See: https://github.com/chzealot/kickstart

# Include sub-config files (optional)
# include:
#   - tools.yaml
#   - repos.yaml

# Dotfiles management (deployed as bare repo to ~/.git)
# dotfiles:
#   repo: git@github.com:yourname/dotfiles.git

# Git repositories (auto clone or pull)
# repos:
#   - url: git@github.com:yourname/project.git
#     path: ~/workspace/project

# Go installation (download from go.dev, fallback to golang.google.cn)
# go: latest

# Tools to install (brew on macOS, auto-detected package manager on Linux)
# tools:
#   - git
#   - curl
#   - jq

# Scripts to run after installation (shell commands)
# scripts:
#   - name: set zsh as default shell
#     run: chsh -s $(which zsh)

# Platform-specific config
# darwin:
#   tools:
#     - coreutils
# linux:
#   tools:
#     - build-essential

# Host-specific config (supports * and ? wildcards)
# hosts:
#   my-macbook:
#     tools:
#       - ffmpeg
#   "dev-*":
#     tools:
#       - docker
`

// Init creates the config directory and a default config file with comments.
func Init(path string) error {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}
		path = filepath.Join(home, defaultConfigDir, defaultConfigFile)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	return os.WriteFile(path, []byte(defaultConfigTemplate), 0644)
}
