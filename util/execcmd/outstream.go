package execcmd

import (
	"io"

	"github.com/moby/term"
	"github.com/sirupsen/logrus"
)

// OutStream writes command output and exposes terminal dimensions.
type OutStream struct {
	CommonStream
	out io.Writer
	le  *logrus.Entry
}

// Write writes output to the underlying writer.
func (o *OutStream) Write(p []byte) (int, error) {
	return o.out.Write(p)
}

// GetTTYSize returns the terminal height and width in characters.
func (o *OutStream) GetTTYSize() (uint, uint) {
	if !o.isTerminal {
		return 0, 0
	}
	ws, err := term.GetWinsize(o.fd)
	if err != nil {
		o.le.WithError(err).Debug("get terminal size")
		if ws == nil {
			return 0, 0
		}
	}
	return uint(ws.Height), uint(ws.Width)
}

// NewOutStream constructs an output stream around a writer.
func NewOutStream(le *logrus.Entry, out io.Writer) io.Writer {
	if out == nil {
		return nil
	}

	fd, isTerminal := term.GetFdInfo(out)
	return &OutStream{
		CommonStream: CommonStream{fd: fd, isTerminal: isTerminal},
		out:          out,
		le:           le,
	}
}
