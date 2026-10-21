package shell

import (
	"context"
	"os"
	osSignal "os/signal"

	"github.com/moby/moby/client"
	"github.com/moby/sys/signal"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/util/execcmd"
)

func resizeTTYTo(
	ctx context.Context,
	dockerClient client.APIClient,
	id string,
	height uint,
	width uint,
	isExec bool,
) error {
	if height == 0 && width == 0 {
		return nil
	}

	if isExec {
		_, err := dockerClient.ExecResize(ctx, id, client.ExecResizeOptions{Height: height, Width: width})
		return err
	}
	_, err := dockerClient.ContainerResize(ctx, id, client.ContainerResizeOptions{Height: height, Width: width})
	return err
}

// MonitorTTYSize updates a container TTY until ctx ends.
func MonitorTTYSize(
	ctx context.Context,
	le *logrus.Entry,
	dockerClient client.APIClient,
	out *execcmd.OutStream,
	id string,
	isExec bool,
) error {
	resizeTTY := func() error {
		height, width := out.GetTTYSize()
		return resizeTTYTo(ctx, dockerClient, id, height, width, isExec)
	}

	if err := resizeTTY(); err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	osSignal.Notify(sigCh, signal.SIGWINCH)
	go func() {
		defer osSignal.Stop(sigCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				if err := resizeTTY(); err != nil && ctx.Err() == nil {
					le.WithError(err).Debug("resize container TTY")
				}
			}
		}
	}()

	return nil
}
