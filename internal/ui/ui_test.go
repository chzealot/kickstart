package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestTitle(t *testing.T) {
	output := captureStdout(func() { Title("Hello") })
	if !strings.Contains(output, "Hello") {
		t.Errorf("Title output should contain 'Hello', got %q", output)
	}
}

func TestSuccess(t *testing.T) {
	output := captureStdout(func() { Success("done %s", "ok") })
	if !strings.Contains(output, "done ok") {
		t.Errorf("Success output should contain 'done ok', got %q", output)
	}
}

func TestError(t *testing.T) {
	output := captureStdout(func() { Error("failed %d", 1) })
	if !strings.Contains(output, "failed 1") {
		t.Errorf("Error output should contain 'failed 1', got %q", output)
	}
}

func TestWarn(t *testing.T) {
	output := captureStdout(func() { Warn("caution") })
	if !strings.Contains(output, "caution") {
		t.Errorf("Warn output should contain 'caution', got %q", output)
	}
}

func TestInfo(t *testing.T) {
	output := captureStdout(func() { Info("note") })
	if !strings.Contains(output, "note") {
		t.Errorf("Info output should contain 'note', got %q", output)
	}
}

func TestStep(t *testing.T) {
	output := captureStdout(func() { Step("doing %s", "something") })
	if !strings.Contains(output, "doing something") {
		t.Errorf("Step output should contain 'doing something', got %q", output)
	}
}

func TestDim(t *testing.T) {
	output := captureStdout(func() { Dim("faded text") })
	if !strings.Contains(output, "faded text") {
		t.Errorf("Dim output should contain 'faded text', got %q", output)
	}
}

func TestSection(t *testing.T) {
	output := captureStdout(func() { Section("MySection") })
	if !strings.Contains(output, "MySection") {
		t.Errorf("Section output should contain 'MySection', got %q", output)
	}
}
