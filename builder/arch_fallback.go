// +build !linux

package builder

import (
	log "github.com/Sirupsen/logrus"
	"github.com/paralin/scratchbuild/arch"
	"runtime"
	"syscall"
)

// detectArch attempts to detect the arch
func detectArch() arch.KnownArch {
	a, _ := arch.ParseArch(runtime.GOARCH)
	return a
}
