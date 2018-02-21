package builder

import (
	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"os"
)

// fetchSourceGit attempts to fetch source by git cloning.
func (b *Builder) fetchSourceGit(destination, source string) error {
	le := log.WithField("source", "git")

	le.WithField("url", source).Debug("Cloning")
	_, err := git.PlainClone(destination, false, &git.CloneOptions{
		Progress:          os.Stdout,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		URL:               source,
	})

	return err
}
