package execcmd

import (
	"io"

	"github.com/moby/term"
)

// InStream reads command input and optionally owns the underlying reader.
type InStream struct {
	CommonStream
	in    io.Reader
	close bool
}

// Read reads from the underlying input.
func (i *InStream) Read(p []byte) (int, error) {
	if i.in == nil {
		return 0, io.EOF
	}
	return i.in.Read(p)
}

// Close closes the underlying reader when ownership was requested.
func (i *InStream) Close() error {
	closer, ok := i.in.(io.ReadCloser)
	if !i.close || !ok {
		return nil
	}
	return closer.Close()
}

// IsTTY reports whether the input is a terminal.
func (i *InStream) IsTTY() bool {
	return i.isTerminal
}

// NewInStream constructs an input stream around a reader.
func NewInStream(in io.Reader, closeReader bool) io.ReadCloser {
	if in == nil {
		return nil
	}
	fd, isTerminal := term.GetFdInfo(in)
	return &InStream{
		CommonStream: CommonStream{fd: fd, isTerminal: isTerminal},
		in:           in,
		close:        closeReader,
	}
}
