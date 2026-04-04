package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const defaultConfigName = ".kickstart"

// Section holds configuration for a scope (global, platform, or host).
type Section struct {
	Dotfiles *DotfilesConfig `yaml:"dotfiles,omitempty"`
	Repos    []RepoConfig    `yaml:"repos,omitempty"`
	Tools    []string        `yaml:"tools,omitempty"`
	Configs  []ConfigTask    `yaml:"configs,omitempty"`
}

// rawConfig is the full YAML structure before merging.
type rawConfig struct {
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
// Merge order: global → platform (darwin/linux/windows) → matching hosts.
// If path is empty, it defaults to ~/.kickstart.
// Returns an empty config (not an error) if the file does not exist.
func Load(path string) (*Config, error) {
	return loadWithEnv(path, runtime.GOOS, hostname())
}

// loadWithEnv is the testable core of Load.
func loadWithEnv(path, goos, host string) (*Config, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("获取用户主目录失败: %w", err)
		}
		path = filepath.Join(home, defaultConfigName)
	}

	cfg := &Config{Path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
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
		return "~/" + defaultConfigName
	}
	return filepath.Join(home, defaultConfigName)
}

// Exists reports whether the config file exists.
func (c *Config) Exists() bool {
	_, err := os.Stat(c.Path)
	return err == nil
}
