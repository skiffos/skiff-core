package arch

import (
	"regexp"
)

// KnownArch represents a known architecture.
type KnownArch int

const (
	NONE KnownArch = iota
	// AMD64 architecture.
	AMD64
	// ARM architecture.
	ARM
	// ARM64 64-bit ARM architecture.
	ARM64
)

// DefaultArch is the architecture most Docker images are compatible with on default.
var DefaultArch KnownArch = AMD64

// KnownArchNames are the string regex representations of KnownArch.
var KnownArchNames = map[string]KnownArch{
	"arm":     ARM,
	"armv*":   ARM,
	"aarch64": ARM64,

	"x86_64": AMD64,
	"amd64":  AMD64,
	"i386":   AMD64, // iffy
}

// KnownArchCompat contains known compatibilities between architectures.
var KnownArchCompat = map[KnownArch][]KnownArch{
	// ARM64 can use ARM images.
	ARM64: {ARM},
}

// ParseArch attempts to determine which arch the (uname -m) output represents.
func ParseArch(arch string) (KnownArch, bool) {
	for archp, archv := range KnownArchNames {
		matched, _ := regexp.MatchString(archp, arch)
		if !matched {
			continue
		}
		return archv, true
	}
	return AMD64, false
}
