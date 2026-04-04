package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupConfigDir creates a ~/.kickstart/ style directory with config files.
func setupConfigDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte(content), 0644)
	}
	return dir
}

func TestLoad_FullConfig(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
dotfiles:
  repo: git@github.com:user/dotfiles.git

repos:
  - url: git@github.com:user/project.git
    path: ~/workspace/project

tools:
  - rsync
  - git
  - jq

configs:
  - name: zsh 默认 shell
    run: chsh -s $(which zsh)
`,
	})

	cfg, err := loadWithEnv(dir, "darwin", "myhost")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Dotfiles == nil || cfg.Dotfiles.Repo != "git@github.com:user/dotfiles.git" {
		t.Fatalf("unexpected dotfiles: %+v", cfg.Dotfiles)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(cfg.Repos))
	}
	if len(cfg.Tools) != 3 {
		t.Fatalf("expected 3 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	if len(cfg.Configs) != 1 {
		t.Fatalf("expected 1 config task, got %d", len(cfg.Configs))
	}
}

func TestLoad_Include(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
include:
  - tools.yaml
  - repos.yaml

dotfiles:
  repo: git@github.com:user/dotfiles.git
`,
		"tools.yaml": `
tools:
  - git
  - curl
  - jq
`,
		"repos.yaml": `
repos:
  - url: git@github.com:user/project.git
    path: ~/workspace/project
`,
	})

	cfg, err := loadWithEnv(dir, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Dotfiles == nil || cfg.Dotfiles.Repo != "git@github.com:user/dotfiles.git" {
		t.Fatalf("unexpected dotfiles: %+v", cfg.Dotfiles)
	}
	if len(cfg.Tools) != 3 {
		t.Fatalf("expected 3 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(cfg.Repos))
	}
}

func TestLoad_IncludeWithPlatform(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
include:
  - darwin.yaml

tools:
  - git
`,
		"darwin.yaml": `
darwin:
  tools:
    - coreutils
  repos:
    - url: git@github.com:user/mac.git
      path: ~/mac
`,
	})

	// Test darwin
	cfg, err := loadWithEnv(dir, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("darwin: expected 2 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("darwin: expected 1 repo, got %d", len(cfg.Repos))
	}

	// Test linux - should not get darwin tools
	cfg, err = loadWithEnv(dir, "linux", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 1 {
		t.Fatalf("linux: expected 1 tool, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
}

func TestLoad_IncludeWithHosts(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
include:
  - hosts.yaml

tools:
  - git
`,
		"hosts.yaml": `
hosts:
  my-macbook:
    tools:
      - ffmpeg
  "dev-*":
    tools:
      - docker
`,
	})

	cfg, err := loadWithEnv(dir, "darwin", "my-macbook")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}

	cfg, err = loadWithEnv(dir, "linux", "dev-server1")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 || cfg.Tools[1] != "docker" {
		t.Fatalf("expected [git docker], got %v", cfg.Tools)
	}
}

func TestLoad_NestedInclude(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
include:
  - base.yaml
tools:
  - git
`,
		"base.yaml": `
include:
  - extra.yaml
tools:
  - curl
`,
		"extra.yaml": `
tools:
  - wget
`,
	})

	cfg, err := loadWithEnv(dir, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// git (main) + curl (base) + wget (extra)
	if len(cfg.Tools) != 3 {
		t.Fatalf("expected 3 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
}

func TestLoad_CircularInclude(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
include:
  - a.yaml
`,
		"a.yaml": `
include:
  - config.yaml
tools:
  - git
`,
	})

	_, err := loadWithEnv(dir, "darwin", "")
	if err == nil {
		t.Fatal("expected error for circular include")
	}
}

func TestLoad_MissingInclude(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
include:
  - nonexistent.yaml
tools:
  - git
`,
	})

	// Missing includes should be silently skipped
	cfg, err := loadWithEnv(dir, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 1 || cfg.Tools[0] != "git" {
		t.Fatalf("expected [git], got %v", cfg.Tools)
	}
}

func TestLoad_PlatformMerge(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
tools:
  - git

darwin:
  tools:
    - coreutils

linux:
  tools:
    - build-essential
`,
	})

	cfg, err := loadWithEnv(dir, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 || cfg.Tools[1] != "coreutils" {
		t.Errorf("darwin: expected [git coreutils], got %v", cfg.Tools)
	}

	cfg, err = loadWithEnv(dir, "linux", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 || cfg.Tools[1] != "build-essential" {
		t.Errorf("linux: expected [git build-essential], got %v", cfg.Tools)
	}

	cfg, err = loadWithEnv(dir, "windows", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 1 {
		t.Errorf("windows: expected [git], got %v", cfg.Tools)
	}
}

func TestLoad_PlatformDotfilesOverride(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
dotfiles:
  repo: git@github.com:user/dotfiles.git

darwin:
  dotfiles:
    repo: git@github.com:user/mac-dotfiles.git
`,
	})

	cfg, err := loadWithEnv(dir, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Dotfiles.Repo != "git@github.com:user/mac-dotfiles.git" {
		t.Errorf("expected darwin override, got %s", cfg.Dotfiles.Repo)
	}

	cfg, err = loadWithEnv(dir, "linux", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Dotfiles.Repo != "git@github.com:user/dotfiles.git" {
		t.Errorf("expected global dotfiles, got %s", cfg.Dotfiles.Repo)
	}
}

func TestLoad_HostDotfilesOverride(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
dotfiles:
  repo: git@github.com:user/dotfiles.git

darwin:
  dotfiles:
    repo: git@github.com:user/mac-dotfiles.git

hosts:
  special-mac:
    dotfiles:
      repo: git@github.com:user/special-dotfiles.git
`,
	})

	cfg, err := loadWithEnv(dir, "darwin", "special-mac")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Dotfiles.Repo != "git@github.com:user/special-dotfiles.git" {
		t.Errorf("expected host override, got %s", cfg.Dotfiles.Repo)
	}
}

func TestLoad_FullMergeOrder(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": `
include:
  - extra.yaml

tools:
  - git

darwin:
  tools:
    - brew-tool

hosts:
  my-mac:
    tools:
      - host-tool
`,
		"extra.yaml": `
tools:
  - included-tool
`,
	})

	cfg, err := loadWithEnv(dir, "darwin", "my-mac")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// git (main) + included-tool (include) + brew-tool (darwin) + host-tool (host)
	if len(cfg.Tools) != 4 {
		t.Fatalf("expected 4 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": "",
	})

	cfg, err := loadWithEnv(dir, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(cfg.Tools))
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	dir := t.TempDir()

	cfg, err := loadWithEnv(filepath.Join(dir, "config.yaml"), "darwin", "")
	if err != nil {
		t.Fatalf("Load should not error for missing file: %v", err)
	}
	if cfg.Exists() {
		t.Fatal("Exists() should return false")
	}
}

func TestLoad_DefaultPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfgDir := filepath.Join(tmpHome, defaultConfigDir)
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, defaultConfigFile), []byte("tools:\n  - wget\n"), 0644)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) < 1 || cfg.Tools[0] != "wget" {
		t.Fatalf("expected wget, got %v", cfg.Tools)
	}
}

func TestLoad_LegacyFileFallback(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create ~/.kickstart as a file (legacy format)
	legacyPath := filepath.Join(tmpHome, defaultConfigDir)
	os.WriteFile(legacyPath, []byte("tools:\n  - legacy-tool\n"), 0644)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 1 || cfg.Tools[0] != "legacy-tool" {
		t.Fatalf("expected [legacy-tool], got %v", cfg.Tools)
	}
}

func TestLoad_DirectoryPath(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": "tools:\n  - git\n",
	})

	// Passing a directory should load config.yaml inside it
	cfg, err := loadWithEnv(dir, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 1 || cfg.Tools[0] != "git" {
		t.Fatalf("expected [git], got %v", cfg.Tools)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": "invalid: [yaml: broken",
	})

	_, err := loadWithEnv(dir, "darwin", "")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestConfig_Exists(t *testing.T) {
	dir := setupConfigDir(t, map[string]string{
		"config.yaml": "",
	})

	cfg := &Config{Path: filepath.Join(dir, "config.yaml")}
	if !cfg.Exists() {
		t.Fatal("Exists() should return true")
	}

	cfg2 := &Config{Path: filepath.Join(t.TempDir(), "nonexistent")}
	if cfg2.Exists() {
		t.Fatal("Exists() should return false")
	}
}

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "newdir", "config.yaml")

	err := Init(cfgPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// File should exist
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// File should be valid YAML (all commented out → empty config)
	cfg, err := loadWithEnv(cfgPath, "darwin", "")
	if err != nil {
		t.Fatalf("Init created invalid config: %v", err)
	}
	if len(cfg.Tools) != 0 {
		t.Errorf("expected 0 tools in template, got %d", len(cfg.Tools))
	}

	// Content should contain comments
	data, _ := os.ReadFile(cfgPath)
	content := string(data)
	if !strings.Contains(content, "# tools:") {
		t.Error("template missing tools comment")
	}
	if !strings.Contains(content, "# hosts:") {
		t.Error("template missing hosts comment")
	}
}

func TestInit_DefaultPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := Init("")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	expectedPath := filepath.Join(tmpHome, defaultConfigDir, defaultConfigFile)
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("config file not created at default path: %v", err)
	}
}
