//go:build !linux
// +build !linux

package builder

import (
	"runtime"

	"github.com/paralin/scratchbuild/arch"
)

// detectArch attempts to detect the arch
func detectArch() arch.KnownArch {
	a, _ := arch.ParseArch(runtime.GOARCH)
	return a
}
