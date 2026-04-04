package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".kickstart")
	os.WriteFile(cfgPath, []byte(content), 0644)
	return cfgPath
}

func TestLoad_FullConfig(t *testing.T) {
	cfgPath := writeTempConfig(t, `
dotfiles:
  repo: git@github.com:user/dotfiles.git

repos:
  - url: git@github.com:user/project.git
    path: ~/workspace/project
  - url: https://github.com/user/another.git
    path: ~/workspace/another

tools:
  - rsync
  - git
  - jq

configs:
  - name: zsh 默认 shell
    run: chsh -s $(which zsh)
`)

	cfg, err := loadWithEnv(cfgPath, "darwin", "myhost")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Dotfiles == nil || cfg.Dotfiles.Repo != "git@github.com:user/dotfiles.git" {
		t.Fatalf("unexpected dotfiles: %+v", cfg.Dotfiles)
	}
	if len(cfg.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(cfg.Repos))
	}
	if len(cfg.Tools) != 3 {
		t.Fatalf("expected 3 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	if len(cfg.Configs) != 1 {
		t.Fatalf("expected 1 config task, got %d", len(cfg.Configs))
	}
}

func TestLoad_PlatformMerge(t *testing.T) {
	cfgPath := writeTempConfig(t, `
tools:
  - git

darwin:
  tools:
    - coreutils
  repos:
    - url: git@github.com:user/mac.git
      path: ~/mac

linux:
  tools:
    - build-essential
`)

	// Test darwin
	cfg, err := loadWithEnv(cfgPath, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("darwin: expected 2 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	if cfg.Tools[0] != "git" || cfg.Tools[1] != "coreutils" {
		t.Errorf("darwin: unexpected tools: %v", cfg.Tools)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("darwin: expected 1 repo, got %d", len(cfg.Repos))
	}

	// Test linux
	cfg, err = loadWithEnv(cfgPath, "linux", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("linux: expected 2 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	if cfg.Tools[0] != "git" || cfg.Tools[1] != "build-essential" {
		t.Errorf("linux: unexpected tools: %v", cfg.Tools)
	}
	if len(cfg.Repos) != 0 {
		t.Fatalf("linux: expected 0 repos, got %d", len(cfg.Repos))
	}

	// Test windows (no platform section)
	cfg, err = loadWithEnv(cfgPath, "windows", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 1 || cfg.Tools[0] != "git" {
		t.Errorf("windows: expected [git], got %v", cfg.Tools)
	}
}

func TestLoad_PlatformDotfilesOverride(t *testing.T) {
	cfgPath := writeTempConfig(t, `
dotfiles:
  repo: git@github.com:user/dotfiles.git

darwin:
  dotfiles:
    repo: git@github.com:user/mac-dotfiles.git
`)

	cfg, err := loadWithEnv(cfgPath, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Dotfiles == nil || cfg.Dotfiles.Repo != "git@github.com:user/mac-dotfiles.git" {
		t.Errorf("expected darwin dotfiles override, got %+v", cfg.Dotfiles)
	}

	// On linux, should use global dotfiles
	cfg, err = loadWithEnv(cfgPath, "linux", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Dotfiles == nil || cfg.Dotfiles.Repo != "git@github.com:user/dotfiles.git" {
		t.Errorf("expected global dotfiles, got %+v", cfg.Dotfiles)
	}
}

func TestLoad_HostMerge(t *testing.T) {
	cfgPath := writeTempConfig(t, `
tools:
  - git

hosts:
  my-macbook:
    tools:
      - ffmpeg
    repos:
      - url: git@github.com:user/private.git
        path: ~/private
  "dev-*":
    tools:
      - docker
`)

	// Exact match
	cfg, err := loadWithEnv(cfgPath, "darwin", "my-macbook")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(cfg.Repos))
	}

	// Wildcard match
	cfg, err = loadWithEnv(cfgPath, "linux", "dev-server1")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	if cfg.Tools[1] != "docker" {
		t.Errorf("expected docker, got %s", cfg.Tools[1])
	}

	// No match
	cfg, err = loadWithEnv(cfgPath, "darwin", "other-host")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
}

func TestLoad_HostDotfilesOverride(t *testing.T) {
	cfgPath := writeTempConfig(t, `
dotfiles:
  repo: git@github.com:user/dotfiles.git

darwin:
  dotfiles:
    repo: git@github.com:user/mac-dotfiles.git

hosts:
  special-mac:
    dotfiles:
      repo: git@github.com:user/special-dotfiles.git
`)

	// Host overrides platform which overrides global
	cfg, err := loadWithEnv(cfgPath, "darwin", "special-mac")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Dotfiles == nil || cfg.Dotfiles.Repo != "git@github.com:user/special-dotfiles.git" {
		t.Errorf("expected host dotfiles override, got %+v", cfg.Dotfiles)
	}
}

func TestLoad_FullMergeOrder(t *testing.T) {
	cfgPath := writeTempConfig(t, `
tools:
  - git

configs:
  - name: global config
    run: echo global

darwin:
  tools:
    - brew-tool
  configs:
    - name: mac config
      run: echo mac

hosts:
  my-mac:
    tools:
      - special-tool
    configs:
      - name: host config
        run: echo host
`)

	cfg, err := loadWithEnv(cfgPath, "darwin", "my-mac")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Tools: global + darwin + host
	if len(cfg.Tools) != 3 {
		t.Fatalf("expected 3 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	expected := []string{"git", "brew-tool", "special-tool"}
	for i, name := range expected {
		if cfg.Tools[i] != name {
			t.Errorf("tools[%d]: expected %s, got %s", i, name, cfg.Tools[i])
		}
	}

	// Configs: global + darwin + host
	if len(cfg.Configs) != 3 {
		t.Fatalf("expected 3 configs, got %d", len(cfg.Configs))
	}
}

func TestLoad_ToolsOnly(t *testing.T) {
	cfgPath := writeTempConfig(t, `
tools:
  - curl
  - wget
`)

	cfg, err := loadWithEnv(cfgPath, "linux", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Dotfiles != nil {
		t.Fatalf("expected nil dotfiles, got %+v", cfg.Dotfiles)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(cfg.Tools))
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	cfgPath := writeTempConfig(t, "")

	cfg, err := loadWithEnv(cfgPath, "darwin", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(cfg.Tools))
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".kickstart")

	cfg, err := loadWithEnv(cfgPath, "darwin", "")
	if err != nil {
		t.Fatalf("Load should not error for missing file: %v", err)
	}
	if len(cfg.Tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(cfg.Tools))
	}
	if cfg.Exists() {
		t.Fatal("Exists() should return false for missing file")
	}
}

func TestLoad_DefaultPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfgPath := filepath.Join(tmpHome, ".kickstart")
	os.WriteFile(cfgPath, []byte("tools:\n  - wget\n"), 0644)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Tools) < 1 || cfg.Tools[0] != "wget" {
		t.Fatalf("expected wget in tools, got %v", cfg.Tools)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	cfgPath := writeTempConfig(t, "invalid: [yaml: broken")

	_, err := loadWithEnv(cfgPath, "darwin", "")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestConfig_Exists(t *testing.T) {
	cfgPath := writeTempConfig(t, "")

	cfg := &Config{Path: cfgPath}
	if !cfg.Exists() {
		t.Fatal("Exists() should return true")
	}

	cfg2 := &Config{Path: filepath.Join(t.TempDir(), "nonexistent")}
	if cfg2.Exists() {
		t.Fatal("Exists() should return false")
	}
}
