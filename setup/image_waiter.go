package setup

import (
	"io"
)

// ImageWaiter can wait for an image to complete.
type ImageWaiter interface {
	WaitForImage(ref string, logOutput io.Writer) error
}
