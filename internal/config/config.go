package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultConfigName = ".kickstart"

// Config holds the parsed configuration.
type Config struct {
	Path     string          `yaml:"-"`
	Dotfiles *DotfilesConfig `yaml:"dotfiles,omitempty"`
	Tools    []string        `yaml:"tools,omitempty"`
	Configs  []ConfigTask    `yaml:"configs,omitempty"`
}

// DotfilesConfig holds dotfiles repository settings.
type DotfilesConfig struct {
	Repo string `yaml:"repo"`
}

// ConfigTask holds a named shell command to run.
type ConfigTask struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

// Load reads and parses the config file.
// If path is empty, it defaults to ~/.kickstart.
// Returns an empty config (not an error) if the file does not exist.
func Load(path string) (*Config, error) {
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

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfg.Path = path
	return cfg, nil
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
