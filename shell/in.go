package shell

import (
	"io"

	"github.com/docker/docker/pkg/term"
)

// InStream is an input stream used by the DockerCli to read user input
type InStream struct {
	CommonStream
	in io.ReadCloser
}

func (i *InStream) Read(p []byte) (int, error) {
	return i.in.Read(p)
}

// Close implements the Closer interface
func (i *InStream) Close() error {
	return i.in.Close()
}

// IsTty checks if the input is a tty.
func (i *InStream) IsTty() bool {
	return i.isTerminal
}

// NewInStream returns a new InStream object from a ReadCloser
func NewInStream(in io.ReadCloser) *InStream {
	fd, isTerminal := term.GetFdInfo(in)
	return &InStream{CommonStream: CommonStream{fd: fd, isTerminal: isTerminal}, in: in}
}
