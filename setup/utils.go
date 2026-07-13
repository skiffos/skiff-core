package setup

import "os"

func pathToSkiffCore() (string, error) {
	return os.Executable()
}
