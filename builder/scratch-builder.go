package builder

import (
	"context"
	"io"
	"os"
	"strings"

	dockerBuild "github.com/docker/cli/cli/command/image/build"
	buildTypes "github.com/docker/docker/api/types/build"
	imageTypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/paralin/scratchbuild/stack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

// ScratchBuilder builds a rebased scratchbuild image stack.
type ScratchBuilder struct {
	le           *logrus.Entry
	dockerClient client.APIClient
	stack        *stack.ImageStack
	outputStream io.Writer
	forceRemove  bool
}

// NewScratchBuilder constructs a scratchbuild image stack builder.
func NewScratchBuilder(
	le *logrus.Entry,
	imageStack *stack.ImageStack,
	dockerClient client.APIClient,
) *ScratchBuilder {
	return &ScratchBuilder{
		le:           le,
		dockerClient: dockerClient,
		stack:        imageStack,
		outputStream: os.Stdout,
		forceRemove:  true,
	}
}

// SetForceRemove controls removal of intermediate build containers.
func (b *ScratchBuilder) SetForceRemove(forceRemove bool) {
	b.forceRemove = forceRemove
}

// SetOutputStream sets the build progress writer.
func (b *ScratchBuilder) SetOutputStream(output io.Writer) {
	if output == nil {
		b.outputStream = io.Discard
		return
	}
	b.outputStream = output
}

func (b *ScratchBuilder) terminalOutput() (uintptr, bool) {
	file, ok := b.outputStream.(*os.File)
	if !ok {
		return 0, false
	}
	fd := file.Fd()
	return fd, terminal.IsTerminal(int(fd))
}

func (b *ScratchBuilder) pullImage(ctx context.Context, reference string) error {
	outputFD, isTerminal := b.terminalOutput()
	b.le.WithField("ref", reference).Debug("pull scratchbuild base image")
	response, err := b.dockerClient.ImagePull(ctx, reference, imageTypes.PullOptions{})
	if err != nil {
		return errors.Wrap(err, "pull scratchbuild base image")
	}
	defer response.Close()
	return jsonmessage.DisplayJSONMessagesStream(response, b.outputStream, outputFD, isTerminal, nil)
}

func (b *ScratchBuilder) dockerBuild(
	ctx context.Context,
	directory string,
	dockerfileSource string,
	reference string,
) error {
	outputFD, isTerminal := b.terminalOutput()
	const relativeDockerfile = "Dockerfile"
	excludes, err := dockerBuild.ReadDockerignore(directory)
	if err != nil {
		return err
	}
	if err := dockerBuild.ValidateContextDirectory(directory, excludes); err != nil {
		return errors.Wrap(err, "validate scratchbuild context")
	}
	excludes = dockerBuild.TrimBuildFilesFromExcludes(excludes, relativeDockerfile, false)
	buildContext, err := archive.TarWithOptions(directory, &archive.TarOptions{ExcludePatterns: excludes})
	if err != nil {
		return err
	}
	defer buildContext.Close()

	buildContext, dockerfile, err := dockerBuild.AddDockerfileToBuildContext(
		io.NopCloser(strings.NewReader(dockerfileSource)),
		buildContext,
	)
	if err != nil {
		return err
	}
	progressOutput := streamformatter.NewProgressOutput(b.outputStream)
	body := progress.NewProgressReader(
		buildContext,
		progressOutput,
		0,
		"",
		"Sending build context to Docker daemon",
	)
	response, err := b.dockerClient.ImageBuild(ctx, body, buildTypes.ImageBuildOptions{
		PullParent:  false,
		ForceRemove: b.forceRemove,
		Dockerfile:  dockerfile,
		Tags:        []string{reference},
	})
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return jsonmessage.DisplayJSONMessagesStream(response.Body, b.outputStream, outputFD, isTerminal, nil)
}

// Build validates and builds each layer in the image stack.
func (b *ScratchBuilder) Build(ctx context.Context) error {
	if b.stack == nil || len(b.stack.Layers) == 0 {
		return errors.New("scratchbuild image stack is empty")
	}
	for _, layer := range b.stack.Layers[:len(b.stack.Layers)-1] {
		if layer.Dockerfile == nil {
			return errors.Errorf("scratchbuild layer has no source: %s", layer.Reference.String())
		}
	}

	baseLayer := b.stack.Layers[len(b.stack.Layers)-1]
	if baseLayer.Reference.Name() != "scratch" {
		if err := b.pullImage(ctx, baseLayer.Reference.String()); err != nil {
			return err
		}
	}
	for index := len(b.stack.Layers) - 2; index >= 0; index-- {
		layer := b.stack.Layers[index]
		reference := layer.Reference.String()
		b.le.WithField("ref", reference).Debug("build scratchbuild layer")
		dockerfile := layer.ToDockerfile()
		if _, err := io.WriteString(b.outputStream, dockerfile+"\n"); err != nil {
			return err
		}
		if err := b.dockerBuild(ctx, layer.Path, dockerfile, reference); err != nil {
			return err
		}
	}
	return nil
}
