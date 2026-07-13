package setup

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
	"github.com/skiffos/skiff-core/util/execcmd"
)

// Setup coordinates image, container, and host-user setup.
type Setup struct {
	le              *logrus.Entry
	config          *config.Config
	workDir         string
	imageSetups     map[string]*ImageSetup
	containerSetups map[string]*ContainerSetup
	createUsers     bool
}

// SetupJob is a setup operation that can be executed and awaited.
type SetupJob interface {
	Execute(context.Context) error
	Wait(context.Context, io.Writer) error
}

func ensureSlashPrefix(value string) string {
	if !strings.HasPrefix(value, "/") {
		return "/" + value
	}
	return value
}

// WaitForImage waits for a configured image setup to complete.
func (s *Setup) WaitForImage(ctx context.Context, ref string, logOut io.Writer) error {
	imageSetup, ok := s.imageSetups[ref]
	if !ok {
		return errors.Errorf("image is not declared: %s", ref)
	}
	return imageSetup.Wait(ctx, logOut)
}

// WaitForContainer waits for a configured container setup to complete.
func (s *Setup) WaitForContainer(ctx context.Context, name string, logOut io.Writer) (string, error) {
	containerSetup, ok := s.containerSetups[name]
	if !ok {
		return "", errors.Errorf("container is not declared: %s", name)
	}
	return containerSetup.WaitWithID(ctx, logOut)
}

// CheckHasContainer reports whether a container is configured.
func (s *Setup) CheckHasContainer(name string) bool {
	_, ok := s.containerSetups[name]
	return ok
}

// ExecCmdContainer executes a command in a configured Docker container.
func (s *Setup) ExecCmdContainer(
	ctx context.Context,
	containerID string,
	userID string,
	stdIn io.Reader,
	stdOut io.Writer,
	stdErr io.Writer,
	cmd string,
	args ...string,
) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer func() {
		if err := dockerClient.Close(); err != nil {
			s.le.WithError(err).Debug("close Docker client")
		}
	}()

	inspection, err := dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}
	if inspection.State == nil || !inspection.State.Running {
		if err := dockerClient.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
			return errors.Wrap(err, "start container")
		}
	}

	return execcmd.ExecCmdContainer(
		ctx,
		s.le,
		dockerClient,
		containerID,
		userID,
		stdIn,
		stdOut,
		stdErr,
		cmd,
		args...,
	)
}

// NewSetup constructs a setup coordinator.
func NewSetup(le *logrus.Entry, conf *config.Config, workDir string, createUsers bool) *Setup {
	return &Setup{
		le:              le,
		config:          conf,
		workDir:         workDir,
		createUsers:     createUsers,
		imageSetups:     make(map[string]*ImageSetup),
		containerSetups: make(map[string]*ContainerSetup),
	}
}

// Execute runs all configured setup jobs and returns the first error.
func (s *Setup) Execute(ctx context.Context) error {
	var jobs []SetupJob
	addImageJob := func(image *config.ConfigImage) {
		imageSetup := NewImageSetup(s.le, image, s.workDir)
		jobs = append(jobs, imageSetup)
		s.imageSetups[image.Name()] = imageSetup
	}

	for _, image := range s.config.Images {
		addImageJob(image)
	}
	for _, containerConfig := range s.config.Containers {
		if containerConfig.Image != "" {
			if _, ok := s.imageSetups[containerConfig.Image]; !ok {
				image := &config.ConfigImage{}
				image.SetName(containerConfig.Image)
				addImageJob(image)
			}
		}
		containerSetup := NewContainerSetup(s.le, containerConfig, s)
		jobs = append(jobs, containerSetup)
		s.containerSetups[containerConfig.Name()] = containerSetup
	}
	for _, user := range s.config.Users {
		jobs = append(jobs, NewUserSetup(s.le, user, s, s.createUsers))
	}

	results := make(chan error, len(jobs))
	for _, job := range jobs {
		go func(job SetupJob) {
			results <- job.Execute(ctx)
		}(job)
	}

	var firstError error
	for pending := len(jobs); pending > 0; pending-- {
		s.le.Debugf("waiting for %d/%d setup jobs", pending, len(jobs))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-results:
			if err != nil {
				s.le.WithError(err).Error("setup job failed")
				if firstError == nil {
					firstError = err
				}
			}
		}
	}
	return firstError
}

// _ is a type assertion.
var (
	_ ImageWaiter     = ((*Setup)(nil))
	_ ContainerWaiter = ((*Setup)(nil))
)
