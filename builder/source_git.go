package builder

import (
	"context"

	"github.com/go-git/go-git/v5"
)

func (b *Builder) fetchSourceGit(ctx context.Context, destination string, source string) error {
	b.le.WithField("url", source).Debug("clone image source")
	_, err := git.PlainCloneContext(ctx, destination, false, &git.CloneOptions{
		Progress:          b.outputStream,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		URL:               source,
	})
	return err
}
