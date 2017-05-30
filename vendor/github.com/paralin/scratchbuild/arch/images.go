package arch

// ArchBaseImages lists known compatible equivalents for amd64 images on other arches.
var ArchBaseImages = map[KnownArch]map[string]string{
	ARM: {
		"library/ubuntu":    "ioft/armhf-ubuntu",
		"library/alpine":    "container4armhf/armhf-alpine",
		"library/busybox":   "container4armhf/armhf-busybox",
		"library/archlinux": "armv7/armhf-archlinux",
		"library/debian":    "armbuild/debian",
	},
}

// CompatibleBaseImage checks for a equivalent/compatible image for a target platform.
func CompatibleBaseImage(targetArch KnownArch, image string) (string, bool) {
	if targetArch == AMD64 {
		return image, true
	}

	compat := KnownArchCompat[targetArch]
	arches := make([]KnownArch, len(compat)+1)
	arches[0] = targetArch
	copy(arches[1:], compat)

	for _, arch := range arches {
		baseImages, ok := ArchBaseImages[arch]
		if !ok {
			continue
		}
		ti, ok := baseImages[image]
		if !ok {
			continue
		}
		return ti, true
	}
	return "", false
}
