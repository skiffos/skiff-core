package builder

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/skiffos/skiff-core/grsync"
)

func (b *Builder) fetchSourceRsync(ctx context.Context, destination string, source string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.Errorf("image source is not a directory: %s", source)
	}
	b.le.
		WithField("destination", destination).
		WithField("source", source).
		Debug("sync image source")
	if !strings.HasSuffix(destination, "/") {
		destination += "/"
	}
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	task := grsync.NewTask(ctx, source, destination, grsync.RsyncOptions{
		Verbose:   true,
		Archive:   true,
		Recursive: true,
		Links:     true,
	})
	return task.Run()
}
