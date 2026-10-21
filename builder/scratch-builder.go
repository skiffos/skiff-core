package builder

import (
	"context"
	"io"
	"os"

	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/jsonmessage"
	"github.com/paralin/scratchbuild/stack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

func (b *ScratchBuilder) pullImage(ctx context.Context, reference string) error {
	b.le.WithField("ref", reference).Debug("pull scratchbuild base image")
	response, err := b.dockerClient.ImagePull(ctx, reference, client.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "pull scratchbuild base image")
	}
	defer response.Close()
	outputFD, isTerminal := terminalOutput(b.outputStream)
	return jsonmessage.DisplayJSONMessagesStream(response, b.outputStream, outputFD, isTerminal, nil)
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
		options := client.ImageBuildOptions{ForceRemove: b.forceRemove, Tags: []string{reference}}
		if err := dockerBuild(ctx, b.dockerClient, b.outputStream, layer.Path, "Dockerfile", dockerfile, options); err != nil {
			return err
		}
	}
	return nil
}
