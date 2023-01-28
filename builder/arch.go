//go:build linux
// +build linux

package builder

import (
	"runtime"
	"syscall"

	"github.com/paralin/scratchbuild/arch"
	log "github.com/sirupsen/logrus"
)

// detectMachineId returns uname -m or GOARCH
func detectMachineId() string {
	un := &syscall.Utsname{}
	mname := runtime.GOARCH
	if err := syscall.Uname(un); err != nil {
		log.WithError(err).Warn("Unable to detect arch via uname, using GOARCH.")
	} else {
		var data []byte
		for _, byt := range un.Machine[:] {
			if byt == 0 {
				break
			}
			data = append(data, byte(byt))
		}
		mname = string(data)
	}
	return mname
}

// detectArch attempts to detect the arch
func detectArch() arch.KnownArch {
	uname := detectMachineId()
	arc, _ := arch.ParseArch(uname)
	return arc
}
