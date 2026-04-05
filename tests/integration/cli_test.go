package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := t.TempDir() + "/kickstart"
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = "../../"
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binary
}

func runKickstart(t *testing.T, binary string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// createTempConfig creates a temporary config file and returns the path.
func createTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".kickstart")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return cfgPath
}

func TestCLI_Help(t *testing.T) {
	bin := buildBinary(t)
	out, err := runKickstart(t, bin, "--help")
	if err != nil {
		t.Fatalf("--help failed: %v\n%s", err, out)
	}

	expected := []string{
		"一键初始化新电脑环境",
		"run",
		"dotfiles",
		"go",
		"install",
		"repos",
		"scripts",
		"status",
		"upgrade",
		"--config",
		"--dry-run",
		"--verbose",
		"--version",
	}
	for _, s := range expected {
		if !strings.Contains(out, s) {
			t.Errorf("help output missing %q", s)
		}
	}
}

func TestCLI_Version(t *testing.T) {
	bin := buildBinary(t)
	out, err := runKickstart(t, bin, "--version")
	if err != nil {
		t.Fatalf("--version failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "kickstart version") {
		t.Errorf("version output = %q, expected to contain 'kickstart version'", out)
	}
}

func TestCLI_Run(t *testing.T) {
	bin := buildBinary(t)
	cfg := createTempConfig(t, "tools:\n  - git\n")
	out, err := runKickstart(t, bin, "run", "-c", cfg)
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "初始化完成") {
		t.Errorf("run output missing '初始化完成', got %q", out)
	}
}

func TestCLI_Run_NoConfig(t *testing.T) {
	bin := buildBinary(t)
	out, err := runKickstart(t, bin, "run", "-c", "/tmp/nonexistent_kickstart_config")
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "配置文件不存在") {
		t.Errorf("run output missing '配置文件不存在', got %q", out)
	}
}

func TestCLI_DefaultCommand(t *testing.T) {
	bin := buildBinary(t)
	cfg := createTempConfig(t, "tools:\n  - git\n")
	// Running without subcommand should behave like "run"
	out, err := runKickstart(t, bin, "-c", cfg)
	if err != nil {
		t.Fatalf("default command failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "初始化完成") {
		t.Errorf("default command output missing '初始化完成', got %q", out)
	}
}

func TestCLI_Status(t *testing.T) {
	bin := buildBinary(t)
	out, err := runKickstart(t, bin, "status")
	if err != nil {
		t.Fatalf("status failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "环境状态") {
		t.Errorf("status output missing '环境状态', got %q", out)
	}
}

func TestCLI_Dotfiles(t *testing.T) {
	bin := buildBinary(t)
	cfg := createTempConfig(t, "# empty config\n")
	out, err := runKickstart(t, bin, "dotfiles", "-c", cfg)
	if err != nil {
		t.Fatalf("dotfiles failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "未定义 dotfiles") {
		t.Errorf("dotfiles output missing expected text, got %q", out)
	}
}

func TestCLI_Install(t *testing.T) {
	bin := buildBinary(t)
	cfg := createTempConfig(t, "tools:\n  - git\n")
	out, err := runKickstart(t, bin, "install", "-c", cfg)
	if err != nil {
		t.Fatalf("install failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "安装工具和软件包") {
		t.Errorf("install output missing expected text, got %q", out)
	}
}

func TestCLI_Install_NoConfig(t *testing.T) {
	bin := buildBinary(t)
	out, err := runKickstart(t, bin, "install", "-c", "/tmp/nonexistent_kickstart_config")
	if err != nil {
		t.Fatalf("install failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "配置文件不存在") {
		t.Errorf("install output missing '配置文件不存在', got %q", out)
	}
}

func TestCLI_Scripts(t *testing.T) {
	bin := buildBinary(t)
	out, err := runKickstart(t, bin, "scripts")
	if err != nil {
		t.Fatalf("scripts failed: %v\n%s", err, out)
	}
	// Without config file, should prompt for init (same as other subcommands)
	if !strings.Contains(out, "配置文件不存在") && !strings.Contains(out, "执行配置脚本") {
		t.Errorf("scripts output missing expected text, got %q", out)
	}
}

func TestCLI_InvalidCommand(t *testing.T) {
	bin := buildBinary(t)
	_, err := runKickstart(t, bin, "nonexistent")
	if err == nil {
		t.Error("expected error for invalid command")
	}
}

func TestCLI_StatusWithConfig(t *testing.T) {
	bin := buildBinary(t)
	cfg := createTempConfig(t, "tools:\n  - git\n  - curl\n")
	out, err := runKickstart(t, bin, "status", "-c", cfg)
	if err != nil {
		t.Fatalf("status with -c failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, cfg) {
		t.Errorf("status output should show custom config path, got %q", out)
	}
	if !strings.Contains(out, "git") {
		t.Errorf("status output should show configured tools, got %q", out)
	}
}

func TestCLI_SubcommandHelp(t *testing.T) {
	bin := buildBinary(t)
	subcommands := []string{"run", "dotfiles", "go", "install", "repos", "scripts", "status", "upgrade"}
	for _, sub := range subcommands {
		out, err := runKickstart(t, bin, sub, "--help")
		if err != nil {
			t.Errorf("%s --help failed: %v\n%s", sub, err, out)
		}
	}
}
