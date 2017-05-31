package stack

import (
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"
)

// ParseImageName parses an image name to a reference.
func ParseImageName(imageName string) (reference.Named, error) {
	ref, err := reference.Parse(imageName)
	if err != nil {
		return nil, err
	}

	named, ok := ref.(reference.Named)
	if !ok {
		return nil, fmt.Errorf("Unable to parse %s: not a named reference", imageName)
	}

	if !strings.Contains(named.Name(), "/") && named.Name() != "scratch" {
		named, err = reference.ParseNamed("docker.io/library/" + imageName)
		if err != nil {
			return nil, err
		}
	}
	return named, nil
}

// ParseNormalizedImageName parses a normalized image name to a reference.
func ParseNormalizedImageName(imageName string) (reference.Named, error) {
	named, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return nil, err
	}
	return named, nil
}
