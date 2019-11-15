package builder

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/client"
	sbbuilder "github.com/paralin/scratchbuild/builder"
	"github.com/paralin/scratchbuild/library"
	"github.com/paralin/scratchbuild/stack"
	"github.com/paralin/skiff-core/config"
)

// Builder manages building images.
type Builder struct {
	lib          *library.LibraryResolver
	config       *config.ConfigImageBuild
	outputStream io.Writer
}

// NewBuilder creates a Builder.
func NewBuilder(config *config.ConfigImageBuild) (*Builder, error) {
	lib, err := globalLibraryCache.GetLibrary()
	if err != nil {
		return nil, err
	}
	return &Builder{config: config, lib: lib}, nil
}

// SetOutputStream sets the output stream.
func (b *Builder) SetOutputStream(s io.Writer) {
	b.outputStream = s
}

// Close the builder to release the resources it's using.
func (b *Builder) Close() {
	if b.lib == nil {
		return
	}

	b.lib = nil
	globalLibraryCache.Release()
}

// Build completes the build process.
func (b *Builder) Build() error {
	tmpDir, err := ioutil.TempDir("", "skiff-core-build-")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	if err := b.fetchSource(tmpDir); err != nil {
		return err
	}

	return b.build(tmpDir)
}

// build completes building the image with a source tree.
func (b *Builder) build(buildPath string) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	arc := detectArch()
	stk, err := stack.ImageStackFromPath(buildPath, b.config.Dockerfile, b.config.ImageName(), b.lib, arc)
	if err != nil {
		return err
	}
	if err := stk.RebaseOnArch(arc); err != nil {
		return err
	}

	bldr := sbbuilder.NewBuilder(stk, dockerClient)
	bldr.SetOutputStream(b.outputStream)
	bldr.SetForceRemove(!b.config.PreserveIntermediate)
	res := make(chan error)
	go func() {
		res <- bldr.Build()
	}()

	time.Sleep(time.Duration(1) * time.Second)

	return <-res
}

// fetchSource downloads the source to a destination path.
func (b *Builder) fetchSource(destination string) error {
	source := b.config.Source

	if source == "" {
		return errors.New("No source specified")
	}

	// determine which kind of URL it is.
	if strings.HasPrefix(source, "git://") ||
		(strings.HasSuffix(source, ".git") && strings.HasPrefix(source, "http")) {
		return b.fetchSourceGit(destination, source)
	}

	if strings.HasSuffix(destination, ".tar.gz") {
		return b.fetchSourceTarball(destination, source)
	}

	if strings.HasPrefix(destination, "/") {
		return b.fetchSourceRsync(destination, source)
	}

	return fmt.Errorf("Unrecognized source kind: %s", destination)
}
