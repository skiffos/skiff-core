//go:build linux

package builder

import (
	"runtime"

	"github.com/paralin/scratchbuild/arch"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func detectMachineID(le *logrus.Entry) string {
	uname := &unix.Utsname{}
	if err := unix.Uname(uname); err != nil {
		le.WithError(err).Warn("detect architecture with uname; using GOARCH")
		return runtime.GOARCH
	}
	machine := make([]byte, 0, len(uname.Machine))
	for _, value := range uname.Machine {
		if value == 0 {
			break
		}
		machine = append(machine, byte(value))
	}
	return string(machine)
}

func detectArch(le *logrus.Entry) arch.KnownArch {
	knownArch, _ := arch.ParseArch(detectMachineID(le))
	return knownArch
}
