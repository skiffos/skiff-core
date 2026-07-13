package config

import (
	"path"
	"strings"
)

// ConfigImageBuild describes how to build a Docker image.
type ConfigImageBuild struct {
	imageName string

	// source identifies a directory, Git repository, or gzip tar archive
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	// root selects a directory inside the source tree
	Root string `json:"root,omitempty" yaml:"root,omitempty"`
	// dockerfile selects a Dockerfile inside the build root
	Dockerfile string `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty"`
	// buildArgs contains Docker build arguments
	BuildArgs map[string]*string `json:"buildArgs,omitempty" yaml:"buildArgs,omitempty"`
	// preserveIntermediate preserves intermediate build containers
	PreserveIntermediate bool `json:"preserveIntermediate,omitempty" yaml:"preserveIntermediate,omitempty"`
	// scratchBuild enables the deprecated architecture-rewrite builder
	ScratchBuild bool `json:"scratchBuild,omitempty" yaml:"scratchBuild,omitempty"`
	// squash requests a single-layer Docker image
	Squash bool `json:"squash,omitempty" yaml:"squash,omitempty"`
}

// ImageName returns the target Docker image reference.
func (c *ConfigImageBuild) ImageName() string {
	return c.imageName
}

// FillDefaults normalizes source-relative build paths.
func (c *ConfigImageBuild) FillDefaults() {
	c.Root = normalizeSourcePath(c.Root)
	c.Dockerfile = normalizeSourcePath(c.Dockerfile)
}

func normalizeSourcePath(value string) string {
	if value == "" {
		return ""
	}
	clean := strings.TrimPrefix(path.Clean("/"+value), "/")
	if clean == "." {
		return ""
	}
	return clean
}
