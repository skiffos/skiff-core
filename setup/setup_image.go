package setup

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/builder"
	"github.com/skiffos/skiff-core/config"
	"github.com/skiffos/skiff-core/util/multiwriter"
)

// ImageSetup pulls or builds one configured image.
type ImageSetup struct {
	le      *logrus.Entry
	logger  multiwriter.MultiWriter
	config  *config.ConfigImage
	workDir string
	done    chan struct{}
	err     error
}

// NewImageSetup constructs an image setup operation.
func NewImageSetup(le *logrus.Entry, conf *config.ConfigImage, workDir string) *ImageSetup {
	setup := &ImageSetup{
		le:      le.WithField("image", conf.Name()),
		config:  conf,
		workDir: workDir,
		done:    make(chan struct{}),
	}
	setup.logger.AddWriter(os.Stdout)
	return setup
}

func (i *ImageSetup) checkImageExists(ctx context.Context, dockerClient client.APIClient, ref string) (bool, error) {
	summaries, err := dockerClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, summary := range summaries {
		for _, name := range summary.RepoTags {
			if name == ref {
				return true, nil
			}
		}
	}
	return false, nil
}

func (i *ImageSetup) pull(ctx context.Context, dockerClient client.APIClient) (pullError error) {
	conf := i.config.Pull
	ref := conf.ImageName()
	if conf.Registry != "" {
		ref = conf.Registry + "/" + ref
	}
	defer func() {
		if pullError != nil {
			i.le.WithError(pullError).WithField("ref", ref).Error("pull image")
		}
	}()

	response, err := dockerClient.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return err
	}
	defer response.Close()
	if err := jsonmessage.DisplayJSONMessagesStream(response, &i.logger, 0, false, nil); err != nil {
		return err
	}
	if conf.Registry != "" {
		if err := dockerClient.ImageTag(ctx, ref, conf.ImageName()); err != nil {
			return err
		}
	}
	return nil
}

func (i *ImageSetup) build(ctx context.Context) (buildError error) {
	defer func() {
		if buildError != nil {
			i.le.WithError(buildError).Error("build image")
		}
	}()

	imageBuilder, err := builder.NewBuilder(i.le, i.config.Build, i.workDir)
	if err != nil {
		return err
	}
	imageBuilder.SetOutputStream(&i.logger)
	return imageBuilder.Build(ctx)
}

// Execute sets up the image and publishes its completion to waiters.
func (i *ImageSetup) Execute(ctx context.Context) (executeError error) {
	defer func() {
		i.err = executeError
		if executeError != nil {
			if _, err := i.logger.Write([]byte("Image setup failed: " + executeError.Error() + "\n")); err != nil {
				i.le.WithError(err).Debug("write image setup error")
			}
		}
		close(i.done)
	}()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer func() {
		if err := dockerClient.Close(); err != nil {
			i.le.WithError(err).Debug("close Docker client")
		}
	}()

	exists, err := i.checkImageExists(ctx, dockerClient, i.config.Name())
	if err != nil {
		return err
	}
	i.le.Debugf("image exists: %v", exists)

	pullConfig := i.config.Pull
	buildConfig := i.config.Build
	if pullConfig == nil && buildConfig == nil {
		if exists {
			return nil
		}
		return errors.Errorf("image not found and no pull or build config is specified: %s", i.config.Name())
	}

	if exists && (pullConfig == nil || pullConfig.Policy != config.ConfigPullPolicyAlways) {
		return nil
	}

	var pullError error
	if pullConfig != nil && pullConfig.Policy != config.ConfigPullPolicyIfBuildFails {
		switch pullConfig.Policy {
		case config.ConfigPullPolicyAlways, config.ConfigPullPolicyIfNotPresent:
			pullError = i.pull(ctx, dockerClient)
			if pullError == nil {
				return nil
			}
		default:
			return errors.Errorf("unknown image pull policy: %s", pullConfig.Policy)
		}
	}

	if buildConfig != nil {
		buildError := i.build(ctx)
		if buildError == nil {
			return nil
		}
		if pullConfig != nil && pullConfig.Policy == config.ConfigPullPolicyIfBuildFails {
			if fallbackError := i.pull(ctx, dockerClient); fallbackError == nil {
				return nil
			} else {
				return errors.Wrapf(fallbackError, "build failed: %v; fallback pull", buildError)
			}
		}
		if pullError != nil {
			return errors.Wrapf(buildError, "pull failed: %v; fallback build", pullError)
		}
		return buildError
	}
	if pullError != nil {
		return pullError
	}
	return errors.Errorf("image cannot be produced: %s", i.config.Name())
}

// Wait waits for image setup completion or context cancellation.
func (i *ImageSetup) Wait(ctx context.Context, logOut io.Writer) error {
	i.logger.AddWriter(logOut)
	defer i.logger.RemoveWriter(logOut)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-i.done:
		return i.err
	}
}

// _ is a type assertion.
var _ SetupJob = ((*ImageSetup)(nil))
