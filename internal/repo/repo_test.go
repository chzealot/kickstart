package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSync_Clone(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a local bare repo as source
	srcRepo := filepath.Join(tmpDir, "source.git")
	run(t, "git", "init", "--bare", srcRepo)

	// Create a temp repo, commit something, push to bare
	workRepo := filepath.Join(tmpDir, "work")
	run(t, "git", "clone", srcRepo, workRepo)
	os.WriteFile(filepath.Join(workRepo, "hello.txt"), []byte("hello"), 0644)
	runIn(t, workRepo, "git", "add", ".")
	runIn(t, workRepo, "git", "-c", "user.name=test", "-c", "user.email=test@test.com", "commit", "-m", "init")
	runIn(t, workRepo, "git", "push")

	// Sync should clone
	target := filepath.Join(tmpDir, "target")
	err := Sync(srcRepo, target)
	if err != nil {
		t.Fatalf("Sync (clone) failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(target, "hello.txt")); err != nil {
		t.Fatal("cloned repo missing hello.txt")
	}
}

func TestSync_Pull(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a local bare repo
	srcRepo := filepath.Join(tmpDir, "source.git")
	run(t, "git", "init", "--bare", srcRepo)

	// Clone, commit, push
	workRepo := filepath.Join(tmpDir, "work")
	run(t, "git", "clone", srcRepo, workRepo)
	os.WriteFile(filepath.Join(workRepo, "v1.txt"), []byte("v1"), 0644)
	runIn(t, workRepo, "git", "add", ".")
	runIn(t, workRepo, "git", "-c", "user.name=test", "-c", "user.email=test@test.com", "commit", "-m", "v1")
	runIn(t, workRepo, "git", "push")

	// Clone to target
	target := filepath.Join(tmpDir, "target")
	run(t, "git", "clone", srcRepo, target)

	// Push a new commit from work repo
	os.WriteFile(filepath.Join(workRepo, "v2.txt"), []byte("v2"), 0644)
	runIn(t, workRepo, "git", "add", ".")
	runIn(t, workRepo, "git", "-c", "user.name=test", "-c", "user.email=test@test.com", "commit", "-m", "v2")
	runIn(t, workRepo, "git", "push")

	// Sync should pull
	err := Sync(srcRepo, target)
	if err != nil {
		t.Fatalf("Sync (pull) failed: %v", err)
	}

	// Verify new file exists
	if _, err := os.Stat(filepath.Join(target, "v2.txt")); err != nil {
		t.Fatal("pulled repo missing v2.txt")
	}
}

func TestSync_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular directory (not a git repo)
	target := filepath.Join(tmpDir, "notgit")
	os.MkdirAll(target, 0755)
	os.WriteFile(filepath.Join(target, "file.txt"), []byte("x"), 0644)

	err := Sync("https://example.com/repo.git", target)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/workspace", filepath.Join(home, "workspace")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := expandHome(tt.input)
		if got != tt.expected {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Not a git repo
	if isGitRepo(tmpDir) {
		t.Error("expected false for non-git dir")
	}

	// Init a git repo
	run(t, "git", "init", tmpDir)
	if !isGitRepo(tmpDir) {
		t.Error("expected true for git repo")
	}
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func runIn(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v in %s failed: %v\n%s", name, args, dir, err, out)
	}
}
