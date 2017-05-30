package stack

import (
	"strings"

	"github.com/docker/distribution/reference"
)

// ParseImageName parses an image name to a reference.
func ParseImageName(imageName string) (reference.Named, error) {
	named, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(named.Name(), "/") && named.Name() != "scratch" {
		named, err = reference.ParseNormalizedNamed("library/" + imageName)
		if err != nil {
			return nil, err
		}
	}
	return named, nil
}
