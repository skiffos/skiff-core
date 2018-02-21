package builder

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/paralin/skiff-core/builder/rsync"
)

// fetchSourceRsync copies from a local path to the destination.
func (b *Builder) fetchSourceRsync(destination, source string) error {
	st, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("Cannot sync from %s, not a directory.", source)
	}
	log.WithField("source", source).WithField("destination", destination).Debug("Syncing")
	return rsync.CopyDir(source, destination)
}
