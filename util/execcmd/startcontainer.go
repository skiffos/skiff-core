package execcmd

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
)

// StartContainer starts a container and waits for it to be ready.
func StartContainer(
	ctx context.Context,
	dockerClient dockerclient.APIClient,
	containerID string,
	pollDur time.Duration,
) error {
	if pdmin := time.Duration(100) * time.Millisecond; pollDur < pdmin {
		pollDur = pdmin
	}
	err := dockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// wait for the container to start
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-time.After(pollDur):
		}

		ctr, err := dockerClient.ContainerInspect(ctx, containerID)
		if err != nil {
			return err
		}
		if ctr.State == nil {
			continue
		}
		if ctr.State.Dead ||
			(!ctr.State.Running && !ctr.State.Restarting && ctr.State.Status == "exited") {
			return fmt.Errorf("Container failed to start with exit code: %d", ctr.State.ExitCode)
		}

		health := ctr.State.Health
		if health != nil && (health.Status != "none" && health.Status != "healthy") {
			continue
		}

		if ctr.State.Running {
			return nil
		}
	}
}
