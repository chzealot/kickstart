package runner

import (
	"os"
	"os/exec"
	"runtime"
)

// RunShell executes a shell command string with full terminal I/O.
// On Unix uses sh -c, on Windows uses cmd /c.
func RunShell(command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
