package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	sbbuilder "github.com/paralin/scratchbuild/builder"
	"github.com/paralin/scratchbuild/stack"
	"github.com/paralin/skiff-core/config"
	"golang.org/x/crypto/ssh/terminal"
)

// Builder manages building images.
type Builder struct {
	config       *config.ConfigImageBuild
	outputStream io.Writer
}

// NewBuilder creates a Builder.
func NewBuilder(config *config.ConfigImageBuild) (*Builder, error) {
	return &Builder{config: config}, nil
}

// SetOutputStream sets the output stream.
func (b *Builder) SetOutputStream(s io.Writer) {
	b.outputStream = s
}

// Close the builder to release the resources it was using.
func (b *Builder) Close() {}

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

	if b.config.ScratchBuild {
		lib, err := globalLibraryCache.GetLibrary()
		if err != nil {
			return err
		}
		defer globalLibraryCache.Release()

		arc := detectArch()
		stk, err := stack.ImageStackFromPath(buildPath, b.config.Dockerfile, b.config.ImageName(), lib, arc)
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

	if err := b.dockerBuild(dockerClient, buildPath, b.config.ImageName()); err != nil {
		return err
	}

	// race: briefly for image tag to complete
	<-time.After(time.Millisecond * 200)
	return nil
}

// build builds the dockerfile in a directory.
func (b *Builder) dockerBuild(dockerClient client.APIClient, buildPath string, reference string) error {
	isTerminal := false
	var outFd uintptr
	if b.outputStream == os.Stdout {
		outFd = os.Stdout.Fd()
		isTerminal = terminal.IsTerminal(int(outFd))
	}

	relDockerfile := b.config.Dockerfile
	if relDockerfile == "" {
		relDockerfile = "Dockerfile"
	}
	excludes, err := build.ReadDockerignore(buildPath)
	if err != nil {
		return err
	}

	if err := build.ValidateContextDirectory(buildPath, excludes); err != nil {
		return fmt.Errorf("Error with context: %v", err)
	}

	excludes = build.TrimBuildFilesFromExcludes(excludes, relDockerfile, false)
	buildCtx, err := archive.TarWithOptions(buildPath, &archive.TarOptions{
		// Compression:     archive.Gzip, - results in an error
		ExcludePatterns: excludes,
	})
	if err != nil {
		return err
	}

	dockerfilePath := path.Join(buildPath, relDockerfile)
	sourceBin, err := ioutil.ReadFile(dockerfilePath)
	if err != nil {
		return err
	}
	dockerfileSrc := string(sourceBin)

	buildCtx, relDockerfile, err = build.AddDockerfileToBuildContext(&nopCloser{strings.NewReader(dockerfileSrc)}, buildCtx)
	if err != nil {
		return err
	}

	progressOutput := streamformatter.NewProgressOutput(os.Stdout)
	var body io.Reader = progress.NewProgressReader(buildCtx, progressOutput, 0, "", "Sending build context to Docker daemon")
	response, err := dockerClient.ImageBuild(context.Background(), body, types.ImageBuildOptions{
		PullParent:  false,
		ForceRemove: !b.config.PreserveIntermediate,
		Dockerfile:  relDockerfile,
		Tags:        []string{reference},
		Squash:      b.config.Squash,
	})
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return jsonmessage.DisplayJSONMessagesStream(response.Body, b.outputStream, outFd, isTerminal, nil)
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
