package setup

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
	"github.com/skiffos/skiff-core/util/execcmd"
)

// Setup is an instance of a setup process.
type Setup struct {
	config          *config.Config
	workDir         string
	imageSetups     map[string]*ImageSetup
	containerSetups map[string]*ContainerSetup
	createUsers     bool
}

// SetupJob is a setup job that we can wait on.
type SetupJob interface {
	// Execute is a goroutine to execute the job
	Execute() error
	// Wait waits for the job to exit.
	Wait(log io.Writer) error
}

// ensureSlashPrefix ensures a string has a / prefix
func ensureSlashPrefix(orig string) string {
	if !strings.HasPrefix(orig, "/") {
		return "/" + orig
	}
	return orig
}

// WaitForImage waits for a image ref to be ready.
func (s *Setup) WaitForImage(ref string, logger io.Writer) error {
	if setup, ok := s.imageSetups[ref]; ok {
		return setup.Wait(logger)
	}
	return fmt.Errorf("No image %s declared!", ref)
}

// WaitForContainer waits for a container to be ready.
func (s *Setup) WaitForContainer(name string, logOut io.Writer) (string, error) {
	if setup, ok := s.containerSetups[name]; ok {
		return setup.WaitWithId(logOut)
	}
	return "", fmt.Errorf("No container %s declared!", name)
}

// CheckHasContainer checks if there is a container with the specified name.
func (s *Setup) CheckHasContainer(name string) bool {
	_, ok := s.containerSetups[name]
	return ok
}

// ExecCmdContainer executes a command in a container.
func (s *Setup) ExecCmdContainer(containerID, userID string, stdIn io.Reader, stdOut, stdErr io.Writer, cmd string, args ...string) error {
	dockerClient, err := dockerclient.NewEnvClient()
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	// Ensure container is running.
	_ = dockerClient.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})

	return execcmd.ExecCmdContainer(
		context.Background(),
		dockerClient,
		containerID,
		userID,
		stdIn, stdOut, stdErr,
		cmd, args...,
	)
}

// NewSetup builds a new Setup instance.
func NewSetup(conf *config.Config, workDir string, createUsers bool) *Setup {
	return &Setup{
		config:          conf,
		workDir:         workDir,
		createUsers:     createUsers,
		imageSetups:     make(map[string]*ImageSetup),
		containerSetups: make(map[string]*ContainerSetup),
	}
}

// Execute runs the setup process.
func (s *Setup) Execute() error {
	var jobs []SetupJob

	addImageJob := func(image *config.ConfigImage) {
		pend := NewImageSetup(image, s.workDir)
		jobs = append(jobs, pend)
		s.imageSetups[image.Name()] = pend
	}

	for _, image := range s.config.Images {
		addImageJob(image)
	}

	for _, ctr := range s.config.Containers {
		if ctr.Image != "" {
			_, ok := s.imageSetups[ctr.Image]
			if !ok {
				imgJob := &config.ConfigImage{}
				imgJob.SetName(ctr.Image)
				addImageJob(imgJob)
			}
		}
		setup := NewContainerSetup(ctr, s)
		jobs = append(jobs, setup)
		s.containerSetups[ctr.Name()] = setup
	}

	for _, user := range s.config.Users {
		setup := NewUserSetup(user, s, s.createUsers)
		jobs = append(jobs, setup)
	}

	results := make(chan error)
	pendingJobs := len(jobs)
	originalJobs := pendingJobs
	for _, job := range jobs {
		go func(job SetupJob) {
			results <- job.Execute()
		}(job)
	}

	var firstError error = nil
	for pendingJobs > 0 {
		log.Debugf("Waiting for %d/%d jobs...", pendingJobs, originalJobs)
		err := <-results
		if err != nil {
			log.WithError(err).Error("Job error")
			if firstError == nil {
				firstError = err
			}
		}
		pendingJobs--
	}

	return firstError
}
