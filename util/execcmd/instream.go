package execcmd

import (
	"io"

	"github.com/docker/docker/pkg/term"
)

// InStream is an input stream used by the DockerCli to read user input
type InStream struct {
	CommonStream
	in    io.Reader
	close bool
}

func (i *InStream) Read(p []byte) (int, error) {
	return i.in.Read(p)
}

// Close implements the Closer interface
func (i *InStream) Close() error {
	if c, ok := i.in.(io.ReadCloser); i.close && ok {
		return c.Close()
	}
	return nil
}

// IsTty checks if the input is a tty.
func (i *InStream) IsTty() bool {
	return i.isTerminal
}

// NewInStream returns a new InStream object from a ReadCloser
func NewInStream(in io.Reader, close bool) *InStream {
	fd, isTerminal := term.GetFdInfo(in)
	return &InStream{CommonStream: CommonStream{fd: fd, isTerminal: isTerminal}, in: in, close: close}
}
