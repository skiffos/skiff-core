// +build linux

package builder

import (
	log "github.com/Sirupsen/logrus"
	"github.com/paralin/scratchbuild/arch"
	"runtime"
	"syscall"
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
			data = append(data, byte(byt))
			if byt == 0 {
				break
			}
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
