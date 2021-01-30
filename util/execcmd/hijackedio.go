package execcmd

import (
	"io"

	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/sirupsen/logrus"
)

// HijackedIOStreamer handles copying input to and output from streams to the
// connection.
type HijackedIOStreamer struct {
	InputStream  io.ReadCloser
	OutputStream io.Writer
	ErrorStream  io.Writer

	Resp types.HijackedResponse

	Tty bool
}

// Stream handles setting up the IO and then begins streaming stdin/stdout
// to/from the hijacked connection, blocking until it is either done reading
// output, the user inputs the detach key sequence when in TTY mode, or when
// the given context is cancelled.
func (h *HijackedIOStreamer) Stream(ctx context.Context) error {
	outputDone := h.beginOutputStream()
	inputDone, detached := h.beginInputStream()

	select {
	case err := <-outputDone:
		return err
	case <-inputDone:
		// Input stream has closed.
		if h.OutputStream != nil || h.ErrorStream != nil {
			// Wait for output to complete streaming.
			select {
			case err := <-outputDone:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	case err := <-detached:
		// Got a detach key sequence.
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *HijackedIOStreamer) beginOutputStream() <-chan error {
	if h.OutputStream == nil && h.ErrorStream == nil {
		// There is no need to copy output.
		return nil
	}

	outputDone := make(chan error)
	go func() {
		var err error

		// When TTY is ON, use regular copy
		if h.OutputStream != nil {
			if h.Tty {
				_, err = io.Copy(h.OutputStream, h.Resp.Reader)
			} else {
				_, err = stdcopy.StdCopy(h.OutputStream, h.ErrorStream, h.Resp.Reader)
			}
		}

		if err != nil {
			logrus.Debugf("Error receiveStdout: %s", err)
		}

		outputDone <- err
	}()

	return outputDone
}

func (h *HijackedIOStreamer) beginInputStream() (doneC <-chan struct{}, detachedC <-chan error) {
	inputDone := make(chan struct{})
	detached := make(chan error)

	go func() {
		if h.InputStream != nil {
			_, err := io.Copy(h.Resp.Conn, h.InputStream)

			if _, ok := err.(term.EscapeError); ok {
				detached <- err
				return
			}

			if err != nil {
				// This error will also occur on the receive
				// side (from stdout) where it will be
				// propogated back to the caller.
				logrus.Debugf("Error sendStdin: %s", err)
			}
		}

		if err := h.Resp.CloseWrite(); err != nil {
			logrus.Debugf("Couldn't send EOF: %s", err)
		}

		close(inputDone)
	}()

	return inputDone, detached
}
