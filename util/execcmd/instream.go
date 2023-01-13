package execcmd

import (
	"io"

	"github.com/moby/term"
)

// InStream is an input stream used by the DockerCli to read user input
type InStream struct {
	CommonStream
	in    io.Reader
	close bool
}

func (i *InStream) Read(p []byte) (int, error) {
	if i.in == nil {
		return 0, io.EOF
	}
	return i.in.Read(p)
}

// Close implements the Closer interface
func (i *InStream) Close() error {
	if i.in != nil {
		if c, ok := i.in.(io.ReadCloser); i.close && ok {
			return c.Close()
		}
	}
	return nil
}

// IsTty checks if the input is a tty.
func (i *InStream) IsTty() bool {
	return i.isTerminal
}

// NewInStream returns a new InStream object from a ReadCloser
func NewInStream(in io.Reader, close bool) io.ReadCloser {
	if in == nil {
		return nil
	}
	fd, isTerminal := term.GetFdInfo(in)
	return &InStream{CommonStream: CommonStream{fd: fd, isTerminal: isTerminal}, in: in, close: close}
}
