package setup

import (
	"os"
	"path/filepath"

	"github.com/skiffos/skiff-core/util/execcmd"
)

// pathToSkiffCore returns the path to this executable.
func pathToSkiffCore() (string, error) {
	return filepath.Abs(os.Args[0])
}

// execCmd executes a command
var execCmd = execcmd.ExecCmd
