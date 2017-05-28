package setup

import (
	"os"
	"os/exec"
	"path/filepath"
)

// pathToSkiffCore returns the path to this executable.
func pathToSkiffCore() (string, error) {
	return filepath.Abs(os.Args[0])
}

// execCmd executes a command
func execCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
