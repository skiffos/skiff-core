package execcmd

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/moby/term"
	"github.com/sirupsen/logrus"
)

// HijackedIOStreamer copies input and output over a Docker hijacked connection.
type HijackedIOStreamer struct {
	InputStream  io.ReadCloser
	OutputStream io.Writer
	ErrorStream  io.Writer
	Resp         types.HijackedResponse
	TTY          bool

	le *logrus.Entry
}

// NewHijackedIOStreamer constructs a streamer for a Docker hijacked connection.
func NewHijackedIOStreamer(
	le *logrus.Entry,
	resp types.HijackedResponse,
	useTTY bool,
) *HijackedIOStreamer {
	return &HijackedIOStreamer{le: le, Resp: resp, TTY: useTTY}
}

// Stream copies stdin, stdout, and stderr until the command, input, or context ends.
func (h *HijackedIOStreamer) Stream(ctx context.Context) error {
	outputDone := h.beginOutputStream()
	inputDone, detached := h.beginInputStream()

	select {
	case err := <-outputDone:
		return err
	case <-inputDone:
		if h.OutputStream != nil || h.ErrorStream != nil {
			select {
			case err := <-outputDone:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	case err := <-detached:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *HijackedIOStreamer) beginOutputStream() <-chan error {
	if h.OutputStream == nil && h.ErrorStream == nil {
		return nil
	}

	outputDone := make(chan error, 1)
	go func() {
		var err error
		if h.TTY {
			output := h.OutputStream
			if output == nil {
				output = h.ErrorStream
			}
			_, err = io.Copy(output, h.Resp.Reader)
		} else {
			output := h.OutputStream
			if output == nil {
				output = io.Discard
			}
			errorOutput := h.ErrorStream
			if errorOutput == nil {
				errorOutput = io.Discard
			}
			_, err = stdcopy.StdCopy(output, errorOutput, h.Resp.Reader)
		}
		if err != nil {
			h.le.WithError(err).Debug("receive stdout")
		}
		outputDone <- err
	}()
	return outputDone
}

func (h *HijackedIOStreamer) beginInputStream() (<-chan struct{}, <-chan error) {
	inputDone := make(chan struct{})
	detached := make(chan error, 1)
	go func() {
		if h.InputStream != nil {
			_, err := io.Copy(h.Resp.Conn, h.InputStream)
			if _, ok := err.(term.EscapeError); ok {
				detached <- err
				return
			}
			if err != nil {
				// The receive side normally returns the same connection error.
				h.le.WithError(err).Debug("send stdin")
			}
		}
		if err := h.Resp.CloseWrite(); err != nil {
			h.le.WithError(err).Debug("send EOF")
		}
		close(inputDone)
	}()
	return inputDone, detached
}
