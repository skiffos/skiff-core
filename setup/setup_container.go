package setup

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
	"github.com/skiffos/skiff-core/util/multiwriter"
)

// ContainerSetup creates one configured Docker container.
type ContainerSetup struct {
	// le is the logger tagged with the container name.
	le *logrus.Entry
	// config is the container configuration to create.
	config *config.ConfigContainer
	// waiter waits for the container's image to exist.
	waiter ImageWaiter
	// logger copies setup progress to every waiting writer.
	logger multiwriter.MultiWriter
	// done is closed when Execute returns.
	done chan struct{}

	// err is the Execute result, valid once done is closed.
	err error
	// containerID is the created or found container, valid once done is closed.
	containerID string
}

// NewContainerSetup constructs a container setup operation.
func NewContainerSetup(
	le *logrus.Entry,
	conf *config.ConfigContainer,
	waiter ImageWaiter,
) *ContainerSetup {
	return &ContainerSetup{
		le:     le.WithField("container", conf.Name()),
		config: conf,
		waiter: waiter,
		done:   make(chan struct{}),
	}
}

// buildDockerContainer converts the container configuration to Docker create options.
func (c *ContainerSetup) buildDockerContainer() (*container.Config, *container.HostConfig, error) {
	// Describe the process. Docker names carry a leading slash; hostnames must not.
	conf := c.config
	containerConfig := &container.Config{
		Hostname:   strings.TrimPrefix(conf.Name(), "/"),
		Image:      conf.Image,
		Entrypoint: conf.Entrypoint,
		Cmd:        conf.Cmd,
		Env:        conf.Env,
		WorkingDir: conf.WorkingDirectory,
		Tty:        conf.Tty,
		StopSignal: conf.StopSignal,
	}

	// Describe the host resources, mounting tini as init unless disabled.
	useInit := !conf.DisableInit
	hostConfig := &container.HostConfig{
		Binds:       conf.Mounts,
		CapAdd:      conf.CapAdd,
		DNS:         conf.DNS,
		DNSSearch:   conf.DNSSearch,
		ExtraHosts:  conf.Hosts,
		Init:        &useInit,
		Privileged:  conf.Privileged,
		SecurityOpt: conf.SecurityOpt,
		Tmpfs:       conf.TmpFs,
	}
	if useInit {
		hostConfig.Binds = append(hostConfig.Binds, "/usr/bin/tini:/dev/init")
	}
	if conf.RestartPolicy != "" {
		hostConfig.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(conf.RestartPolicy)}
	}

	// Publish each TCP port, letting Docker choose the host port when it is zero.
	for _, portConfig := range conf.Ports {
		if portConfig.HostPort < 0 || portConfig.HostPort > 65535 {
			return nil, nil, errors.Errorf("invalid host TCP port: %d", portConfig.HostPort)
		}
		port, err := nat.NewPort("tcp", strconv.Itoa(portConfig.ContainerPort))
		if err != nil {
			return nil, nil, errors.Wrap(err, "parse container port")
		}
		if containerConfig.ExposedPorts == nil {
			containerConfig.ExposedPorts = make(nat.PortSet)
			hostConfig.PortBindings = make(nat.PortMap)
		}
		containerConfig.ExposedPorts[port] = struct{}{}
		hostPort := ""
		if portConfig.HostPort > 0 {
			hostPort = strconv.Itoa(portConfig.HostPort)
		}
		hostConfig.PortBindings[port] = []nat.PortBinding{{HostPort: hostPort}}
	}

	// Share each requested host namespace.
	if conf.HostNetwork {
		hostConfig.NetworkMode = container.NetworkMode("host")
	}
	if conf.HostIPC {
		hostConfig.IpcMode = container.IpcMode("host")
	}
	if conf.HostPID {
		hostConfig.PidMode = container.PidMode("host")
	}
	if conf.HostUTS {
		hostConfig.UTSMode = container.UTSMode("host")
	}
	return containerConfig, hostConfig, nil
}

// Execute creates or finds the container and publishes completion to waiters.
func (c *ContainerSetup) Execute(ctx context.Context) (executeError error) {
	// Publish the result to waiters when setup returns.
	defer func() {
		c.err = executeError
		close(c.done)
	}()

	// Connect to Docker.
	if c.config.Image == "" {
		return errors.Errorf("container must specify an image: %s", c.config.Name())
	}
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer func() {
		if err := dockerClient.Close(); err != nil {
			c.le.WithError(err).Debug("close Docker client")
		}
	}()

	// Find an existing container by name, before and after its image is ready.
	exists, err := c.findContainer(ctx, dockerClient)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.waiter.WaitForImage(ctx, c.config.Image, &c.logger); err != nil {
			return err
		}
		exists, err = c.findContainer(ctx, dockerClient)
		if err != nil {
			return err
		}
	}

	// Create the container when none exists.
	if !exists {
		containerConfig, hostConfig, err := c.buildDockerContainer()
		if err != nil {
			return err
		}
		result, err := dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, c.config.Name())
		if err != nil {
			return err
		}
		c.containerID = result.ID
		c.le.WithField("id", result.ID).Debug("container created")
		for _, warning := range result.Warnings {
			c.le.Warnf("Docker warning: %s", warning)
		}
	}

	// Start the container when configured to.
	if _, err := c.logger.Write([]byte("Container created/found with ID: " + c.containerID + "\n")); err != nil {
		return errors.Wrap(err, "write container setup log")
	}
	if !c.config.StartAfterCreate {
		return nil
	}
	if _, err := c.logger.Write([]byte("Starting container " + c.containerID + "...\n")); err != nil {
		return errors.Wrap(err, "write container setup log")
	}
	return errors.Wrap(dockerClient.ContainerStart(ctx, c.containerID, container.StartOptions{}), "start container")
}

// findContainer records the ID of the container with the configured name and reports whether it exists.
func (c *ContainerSetup) findContainer(ctx context.Context, dockerClient *client.Client) (bool, error) {
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, err
	}
	for _, candidate := range containers {
		for _, name := range candidate.Names {
			if name == c.config.Name() {
				c.le.Debug("container already exists")
				c.containerID = candidate.ID
				return true, nil
			}
		}
	}
	return false, nil
}

// Wait waits for container setup completion or context cancellation.
func (c *ContainerSetup) Wait(ctx context.Context, logOut io.Writer) error {
	c.logger.AddWriter(logOut)
	defer c.logger.RemoveWriter(logOut)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.err
	}
}

// WaitWithID waits for container setup and returns the container ID.
func (c *ContainerSetup) WaitWithID(ctx context.Context, logOut io.Writer) (string, error) {
	if err := c.Wait(ctx, logOut); err != nil {
		return "", err
	}
	return c.containerID, nil
}

// _ asserts that ContainerSetup is a SetupJob.
var _ SetupJob = (*ContainerSetup)(nil)
