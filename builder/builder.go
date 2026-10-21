package builder

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/moby/client"
	"github.com/paralin/scratchbuild/stack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
)

// Builder builds a configured Docker image from a source tree.
type Builder struct {
	le           *logrus.Entry
	config       *config.ConfigImageBuild
	outputStream io.Writer
	workDir      string
}

// NewBuilder constructs an image builder.
func NewBuilder(
	le *logrus.Entry,
	conf *config.ConfigImageBuild,
	workDir string,
) (*Builder, error) {
	if conf == nil {
		return nil, errors.New("image build config is nil")
	}
	if conf.Source == "" {
		return nil, errors.New("image build source is empty")
	}
	return &Builder{
		le:           le,
		config:       conf,
		outputStream: os.Stdout,
		workDir:      workDir,
	}, nil
}

// SetOutputStream sets the Docker build progress writer.
func (b *Builder) SetOutputStream(output io.Writer) {
	if output == nil {
		b.outputStream = io.Discard
		return
	}
	b.outputStream = output
}

// Build fetches the configured source and builds its Docker image.
func (b *Builder) Build(ctx context.Context) error {
	tempDir, err := os.MkdirTemp(b.workDir, "skiff-core-build-")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			b.le.WithError(err).WithField("path", tempDir).Warn("remove build directory")
		}
	}()

	sourceDir, err := b.fetchSource(ctx, tempDir)
	if err != nil {
		return err
	}
	buildDir := sourceDir
	if b.config.Root != "" {
		root := filepath.Clean(filepath.FromSlash(b.config.Root))
		if !filepath.IsLocal(root) {
			return errors.Errorf("build root is not local: %s", b.config.Root)
		}
		buildDir = filepath.Join(sourceDir, root)
		info, err := os.Stat(buildDir)
		if err != nil {
			return errors.Wrap(err, "inspect build root")
		}
		if !info.IsDir() {
			return errors.Errorf("build root is not a directory: %s", b.config.Root)
		}
	}
	return b.build(ctx, buildDir)
}

func (b *Builder) build(ctx context.Context, buildPath string) error {
	dockerClient, err := client.New(client.FromEnv)
	if err != nil {
		return err
	}
	defer func() {
		if err := dockerClient.Close(); err != nil {
			b.le.WithError(err).Debug("close Docker client")
		}
	}()

	if b.config.ScratchBuild {
		libraryResolver, err := globalLibraryCache.get()
		if err != nil {
			return err
		}
		defer func() {
			if err := globalLibraryCache.release(); err != nil {
				b.le.WithError(err).Warn("remove scratch library cache")
			}
		}()

		targetArch := detectArch(b.le)
		imageStack, err := stack.ImageStackFromPath(
			buildPath,
			b.config.Dockerfile,
			b.config.ImageName(),
			libraryResolver,
			targetArch,
		)
		if err != nil {
			return err
		}
		if err := imageStack.RebaseOnArch(targetArch); err != nil {
			return err
		}

		imageBuilder := NewScratchBuilder(b.le, imageStack, dockerClient)
		imageBuilder.SetOutputStream(b.outputStream)
		imageBuilder.SetForceRemove(!b.config.PreserveIntermediate)
		return imageBuilder.Build(ctx)
	}
	return b.dockerBuild(ctx, dockerClient, buildPath, b.config.ImageName())
}

func (b *Builder) dockerBuild(
	ctx context.Context,
	dockerClient client.APIClient,
	buildPath string,
	reference string,
) error {
	relativeDockerfile := b.config.Dockerfile
	if relativeDockerfile == "" {
		relativeDockerfile = "Dockerfile"
	}
	dockerfileSource, err := os.ReadFile(filepath.Join(buildPath, filepath.FromSlash(relativeDockerfile)))
	if err != nil {
		return err
	}
	return dockerBuild(ctx, dockerClient, b.outputStream, buildPath, relativeDockerfile, string(dockerfileSource), client.ImageBuildOptions{
		ForceRemove: !b.config.PreserveIntermediate,
		Tags:        []string{reference},
		Squash:      b.config.Squash,
		BuildArgs:   b.config.BuildArgs,
	})
}

func (b *Builder) fetchSource(ctx context.Context, destination string) (string, error) {
	source := b.config.Source
	switch {
	case strings.HasPrefix(source, "git://"),
		strings.HasPrefix(source, "ssh://"),
		strings.HasSuffix(source, ".git") && strings.HasPrefix(source, "http"):
		return destination, b.fetchSourceGit(ctx, destination, source)
	case strings.HasSuffix(source, ".tar.gz"), strings.HasSuffix(source, ".tgz"):
		return destination, b.fetchSourceTarball(ctx, destination, source)
	case filepath.IsAbs(source):
		if info, err := os.Stat(source); err == nil && info.IsDir() {
			return source, nil
		}
		return destination, b.fetchSourceRsync(ctx, destination, source)
	default:
		return "", errors.Errorf("unrecognized image source: %s", source)
	}
}
