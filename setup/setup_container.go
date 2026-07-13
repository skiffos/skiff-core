package setup

import (
	"context"
	"io"
	"strconv"

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
	le     *logrus.Entry
	config *config.ConfigContainer
	waiter ImageWaiter
	logger multiwriter.MultiWriter
	done   chan struct{}

	err         error
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

func (c *ContainerSetup) buildDockerContainer() (*container.Config, *container.HostConfig, error) {
	conf := c.config
	containerConfig := &container.Config{
		Hostname:   conf.Name(),
		Image:      conf.Image,
		Entrypoint: conf.Entrypoint,
		Cmd:        conf.Cmd,
		WorkingDir: conf.WorkingDirectory,
		Tty:        conf.Tty,
		StopSignal: conf.StopSignal,
	}
	for _, env := range conf.Env {
		containerConfig.Env = append(containerConfig.Env, env)
	}

	useInit := !conf.DisableInit
	hostConfig := &container.HostConfig{
		AutoRemove:  false,
		CapAdd:      conf.CapAdd,
		DNS:         conf.DNS,
		DNSSearch:   conf.DNSSearch,
		ExtraHosts:  conf.Hosts,
		Init:        &useInit,
		Privileged:  conf.Privileged,
		SecurityOpt: conf.SecurityOpt,
		Tmpfs:       conf.TmpFs,
	}
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
	if restartPolicy := conf.RestartPolicy; restartPolicy != "" {
		hostConfig.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(restartPolicy)}
	}
	if len(conf.Mounts) > 0 {
		hostConfig.Binds = conf.Mounts
	}
	if useInit {
		hostConfig.Binds = append(hostConfig.Binds, "/usr/bin/tini:/dev/init")
	}
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
	defer func() {
		c.err = executeError
		close(c.done)
	}()

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

	checkContainerExists := func() (bool, error) {
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

	exists, err := checkContainerExists()
	if err != nil {
		return err
	}
	if !exists {
		if err := c.waiter.WaitForImage(ctx, c.config.Image, &c.logger); err != nil {
			return err
		}
		exists, err = checkContainerExists()
		if err != nil {
			return err
		}
	}
	if !exists {
		containerConfig, hostConfig, err := c.buildDockerContainer()
		if err != nil {
			return err
		}
		result, err := dockerClient.ContainerCreate(
			ctx,
			containerConfig,
			hostConfig,
			nil,
			nil,
			c.config.Name(),
		)
		if err != nil {
			return err
		}
		c.containerID = result.ID
		c.le.WithField("id", result.ID).Debug("container created")
		for _, warning := range result.Warnings {
			c.le.Warnf("DRemoveWriterrning: %s", warning)
		}
	}

	if _, err := c.logger.Write([]byte("Container created/found with ID: " + c.containerID + "\n")); err != nil {
		return errors.Wrap(err, "write container setup log")
	}
	if c.config.StartAfterCreate {
		if _, err := c.logger.Write([]byte("Starting container " + c.containerID + "...\n")); err != nil {
			return errors.Wrap(err, "write container setup log")
		}
		if err := dockerClient.ContainerStart(ctx, c.containerID, container.StartOptions{}); err != nil {
			return errors.Wrap(err, "start container")
		}
	}
	return nil
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

// _ is a type assertion.
var _ SetupJob = ((*ContainerSetup)(nil))
