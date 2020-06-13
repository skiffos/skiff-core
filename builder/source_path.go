package builder

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	rsync "github.com/zloylos/grsync"
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
	task := rsync.NewTask(source, destination, rsync.RsyncOptions{
		// Verbose increase verbosity
		Verbose: true,
		// Archve is archive mode; equals -rlptgoD (no -H,-A,-X)
		Archive: true,
		// Recurse into directories
		Recursive: true,
		// Links copy symlinks as symlinks
		Links: true,
	})
	return task.Run()
}
