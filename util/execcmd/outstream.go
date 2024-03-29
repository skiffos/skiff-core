package execcmd

import (
	"io"

	"github.com/moby/term"
	"github.com/sirupsen/logrus"
)

// OutStream is an output stream used by the DockerCli to write normal program
// output.
type OutStream struct {
	CommonStream
	out io.Writer
}

func (o *OutStream) Write(p []byte) (int, error) {
	return o.out.Write(p)
}

// GetTtySize returns the height and width in characters of the tty
func (o *OutStream) GetTtySize() (uint, uint) {
	if !o.isTerminal {
		return 0, 0
	}
	ws, err := term.GetWinsize(o.fd)
	if err != nil {
		logrus.Debugf("Error getting size: %s", err)
		if ws == nil {
			return 0, 0
		}
	}
	return uint(ws.Height), uint(ws.Width)
}

// NewOutStream returns a new OutStream object from a Writer
func NewOutStream(out io.Writer) io.Writer {
	if out == nil {
		return nil
	}

	fd, isTerminal := term.GetFdInfo(out)
	return &OutStream{CommonStream: CommonStream{fd: fd, isTerminal: isTerminal}, out: out}
}
