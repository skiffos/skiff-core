package execcmd

import (
	"os"
	"os/exec"
)

// ExecCmd executes a command on the local machine.
func ExecCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
