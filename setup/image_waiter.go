package setup

import (
	"context"
	"io"
)

// ImageWaiter waits for an image setup to complete.
type ImageWaiter interface {
	WaitForImage(context.Context, string, io.Writer) error
}
