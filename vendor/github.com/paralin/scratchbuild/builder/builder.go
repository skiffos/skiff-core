package builder

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/paralin/scratchbuild/stack"
	"golang.org/x/crypto/ssh/terminal"

	log "github.com/sirupsen/logrus"
)

// Builder manages the Docker Image build process.
type Builder struct {
	dockerClient client.APIClient
	stack        *stack.ImageStack
}

// NewBuilder creates a new builder.
func NewBuilder(stack *stack.ImageStack, dockerClient client.APIClient) *Builder {
	return &Builder{
		stack:        stack,
		dockerClient: dockerClient,
	}
}

// nopCloser wraps readers without a Close()
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func (b *Builder) pullImage(baseRef string) error {
	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	log.WithField("ref", baseRef).Debug("Pulling")
	rc, err := b.dockerClient.ImagePull(context.Background(), baseRef, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("Unable to pull image: %v", err)
	}
	defer rc.Close()
	return jsonmessage.DisplayJSONMessagesStream(rc, os.Stdout, os.Stdout.Fd(), isTerminal, nil)
}

// build builds the dockerfile in a directory.
func (b *Builder) dockerBuild(dir string, dockerfileSrc string, reference string) error {
	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	relDockerfile := "Dockerfile"
	excludes, err := build.ReadDockerignore(dir)
	if err != nil {
		return err
	}

	if err := build.ValidateContextDirectory(dir, excludes); err != nil {
		return fmt.Errorf("Error with context: %v", err)
	}

	excludes = build.TrimBuildFilesFromExcludes(excludes, relDockerfile, false)
	buildCtx, err := archive.TarWithOptions(dir, &archive.TarOptions{
		// Compression:     archive.Gzip, - results in an error
		ExcludePatterns: excludes,
	})
	if err != nil {
		return err
	}

	buildCtx, relDockerfile, err = build.AddDockerfileToBuildContext(&nopCloser{strings.NewReader(dockerfileSrc)}, buildCtx)
	if err != nil {
		return err
	}

	progressOutput := streamformatter.NewProgressOutput(os.Stdout)
	var body io.Reader = progress.NewProgressReader(buildCtx, progressOutput, 0, "", "Sending build context to Docker daemon")
	response, err := b.dockerClient.ImageBuild(context.Background(), body, types.ImageBuildOptions{
		PullParent: false,
		Dockerfile: relDockerfile,
		Tags:       []string{reference},
	})
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, os.Stdout.Fd(), isTerminal, nil)
}

// Build traverses the stack and builds the images.
func (b *Builder) Build() error {
	// Ensure we have a clear path to our destination.
	// This means our chain needs to have Dockerfiles all the way down to the last element.
	for _, layer := range b.stack.Layers[:len(b.stack.Layers)-1] {
		layerRef := layer.Reference.String()
		if layer.Dockerfile == nil {
			return fmt.Errorf("Cannot build, do not know the source for %s", layerRef)
		}
	}

	// Check if we need to pull the final element.
	baseLayer := b.stack.Layers[len(b.stack.Layers)-1]
	if baseLayer.Reference.Name() != "scratch" {
		baseRef := baseLayer.Reference.String()
		if err := b.pullImage(baseRef); err != nil {
			return err
		}
	}

	// Build the images
	for i := len(b.stack.Layers) - 2; i >= 0; i-- {
		layer := b.stack.Layers[i]

		layerRef := layer.Reference.String()
		log.WithField("ref", layerRef).Debug("Building")
		dockerSrc := layer.ToDockerfile()
		fmt.Printf("%s\n", dockerSrc)
		if err := b.dockerBuild(layer.Path, dockerSrc, layerRef); err != nil {
			return err
		}
	}

	return nil
}
