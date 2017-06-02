package arch

import (
	"strings"
)

// ArchBaseImages lists known compatible equivalents for amd64 images on other arches.
var ArchBaseImages = map[KnownArch]map[string]string{
	ARMV8: {
		"library/debian":    "arm64v8/debian",
		"library/alpine":    "arm64v8/alpine",
		"library/php":       "arm64v8/php",
		"library/ubuntu":    "arm64v8/ubuntu",
		"library/wordpress": "arm64v8/wordpress",
		"library/busybox":   "arm64v8/busybox",
		"library/ruby":      "arm64v8/ruby",
		"library/httpd":     "arm64v8/httpd",
		"library/fedora":    "arm64v8/fedora",
	},
	ARMV7: {
		"library/opensuse":  "arm32v7/opensuse",
		"library/ubuntu":    "arm32v7/ubuntu",
		"library/alpine":    "container4armhf/armhf-alpine",
		"library/busybox":   "container4armhf/armhf-busybox",
		"library/archlinux": "armv7/armhf-archlinux",
		"library/debian":    "armbuild/debian",
	},
	ARMV6: {
		"library/alpine":   "arm32v6/alpine",
		"library/openjdk":  "arm32v6/openjdk",
		"library/tomcat":   "arm32v6/tomcat",
		"library/bash":     "arm32v6/bash",
		"library/golang":   "arm32v6/golang",
		"library/postgres": "arm32v6/postgres",
		"library/haproxy":  "arm32v6/haproxy",
		"library/busybox":  "arm32v6/busybox",
	},
}

// CompatibleBaseImage checks for a equivalent/compatible image for a target platform.
func CompatibleBaseImage(targetArch KnownArch, image string) (string, bool) {
	if targetArch == AMD64 {
		return image, true
	}

	if strings.HasPrefix(image, "docker.io/") {
		image = image[len("docker.io/"):]
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
