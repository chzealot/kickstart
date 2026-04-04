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
	Repos    []RepoConfig    `yaml:"repos,omitempty"`
	Tools    []string        `yaml:"tools,omitempty"`
	Configs  []ConfigTask    `yaml:"configs,omitempty"`
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
	Repos    []RepoConfig
	Tools    []string
	Configs  []ConfigTask
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

// ConfigTask holds a named shell command to run.
type ConfigTask struct {
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
		Repos:    append([]RepoConfig{}, raw.Section.Repos...),
		Tools:    append([]string{}, raw.Section.Tools...),
		Configs:  append([]ConfigTask{}, raw.Section.Configs...),
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
	cfg.Repos = merged.Repos
	cfg.Tools = merged.Tools
	cfg.Configs = merged.Configs
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
// Tools, Repos, Configs are appended. Dotfiles is overridden if set.
func mergeSection(dst, src *Section) {
	if src.Dotfiles != nil {
		dst.Dotfiles = src.Dotfiles
	}
	dst.Repos = append(dst.Repos, src.Repos...)
	dst.Tools = append(dst.Tools, src.Tools...)
	dst.Configs = append(dst.Configs, src.Configs...)
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

const defaultConfigTemplate = `# kickstart 配置文件
# 详细说明请参考: https://github.com/chzealot/kickstart

# 引入子配置文件（可选）
# include:
#   - tools.yaml
#   - repos.yaml

# Dotfiles 管理（bare repo 方式部署到 ~/.git）
# dotfiles:
#   repo: git@github.com:yourname/dotfiles.git

# Git 仓库列表（自动 clone 或 pull 更新）
# repos:
#   - url: git@github.com:yourname/project.git
#     path: ~/workspace/project

# 工具安装（macOS 用 brew，Linux 自动检测包管理器）
# tools:
#   - git
#   - curl
#   - jq

# 软件配置（安装完成后执行的 shell 命令）
# configs:
#   - name: zsh 默认 shell
#     run: chsh -s $(which zsh)

# 平台特定配置
# darwin:
#   tools:
#     - coreutils
# linux:
#   tools:
#     - build-essential

# 主机名特定配置（支持 * 和 ? 通配符）
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
