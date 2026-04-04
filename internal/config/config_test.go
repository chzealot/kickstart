package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FullConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".kickstart")

	content := `
dotfiles:
  repo: git@github.com:user/dotfiles.git

tools:
  - rsync
  - git
  - jq

configs:
  - name: zsh 默认 shell
    run: chsh -s $(which zsh)
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Dotfiles == nil || cfg.Dotfiles.Repo != "git@github.com:user/dotfiles.git" {
		t.Fatalf("unexpected dotfiles: %+v", cfg.Dotfiles)
	}

	if len(cfg.Tools) != 3 {
		t.Fatalf("expected 3 tools, got %d: %v", len(cfg.Tools), cfg.Tools)
	}
	expected := []string{"rsync", "git", "jq"}
	for i, name := range expected {
		if cfg.Tools[i] != name {
			t.Errorf("tools[%d]: expected %s, got %s", i, name, cfg.Tools[i])
		}
	}

	if len(cfg.Configs) != 1 {
		t.Fatalf("expected 1 config task, got %d", len(cfg.Configs))
	}
	if cfg.Configs[0].Name != "zsh 默认 shell" {
		t.Errorf("unexpected config name: %s", cfg.Configs[0].Name)
	}
}

func TestLoad_ToolsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".kickstart")

	content := `
tools:
  - curl
  - wget
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Dotfiles != nil {
		t.Fatalf("expected nil dotfiles, got %+v", cfg.Dotfiles)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(cfg.Tools))
	}
	if len(cfg.Configs) != 0 {
		t.Fatalf("expected 0 configs, got %d", len(cfg.Configs))
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".kickstart")
	os.WriteFile(cfgPath, []byte(""), 0644)

	cfg, err := Load(cfgPath)
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

	cfg, err := Load(cfgPath)
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
	if len(cfg.Tools) != 1 || cfg.Tools[0] != "wget" {
		t.Fatalf("expected [wget], got %v", cfg.Tools)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".kickstart")
	os.WriteFile(cfgPath, []byte("invalid: [yaml: broken"), 0644)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestConfig_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".kickstart")
	os.WriteFile(cfgPath, []byte(""), 0644)

	cfg := &Config{Path: cfgPath}
	if !cfg.Exists() {
		t.Fatal("Exists() should return true")
	}

	cfg2 := &Config{Path: filepath.Join(tmpDir, "nonexistent")}
	if cfg2.Exists() {
		t.Fatal("Exists() should return false")
	}
}
