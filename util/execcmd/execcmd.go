package execcmd

import (
	"context"
	"os"
	"os/exec"
)

// ExecCmd executes a local command with inherited standard output and error.
func ExecCmd(ctx context.Context, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
