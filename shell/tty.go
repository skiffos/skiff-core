package shell

import (
	"os"
	gosignal "os/signal"

	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/signal"
	"github.com/skiffos/skiff-core/util/execcmd"
)

// resizeTtyTo resizes tty to specific height and width
func resizeTtyTo(ctx context.Context, client client.ContainerAPIClient, id string, height, width uint, isExec bool) {
	if height == 0 && width == 0 {
		return
	}

	options := types.ResizeOptions{
		Height: height,
		Width:  width,
	}

	var err error
	if isExec {
		err = client.ContainerExecResize(ctx, id, options)
	} else {
		err = client.ContainerResize(ctx, id, options)
	}

	_ = err // Ignore this error for now.
	/*
		if err != nil {
		}
	*/
}

// MonitorTtySize updates the container tty size when the terminal tty changes size
func MonitorTtySize(ctx context.Context, client client.APIClient, out *execcmd.OutStream, id string, isExec bool) error {
	resizeTty := func() {
		height, width := out.GetTtySize()
		resizeTtyTo(ctx, client, id, height, width, isExec)
	}

	resizeTty()

	sigchan := make(chan os.Signal, 1)
	gosignal.Notify(sigchan, signal.SIGWINCH)
	go func() {
		for range sigchan {
			resizeTty()
		}
	}()

	return nil
}
