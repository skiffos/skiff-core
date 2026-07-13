//go:build !linux

package builder

import (
	"runtime"

	"github.com/paralin/scratchbuild/arch"
	"github.com/sirupsen/logrus"
)

func detectArch(_ *logrus.Entry) arch.KnownArch {
	knownArch, _ := arch.ParseArch(runtime.GOARCH)
	return knownArch
}
