package execcmd

import "github.com/moby/term"

// CommonStream tracks terminal state shared by input and output streams.
type CommonStream struct {
	fd         uintptr
	isTerminal bool
	state      *term.State
}

// FD returns the stream file descriptor.
func (s *CommonStream) FD() uintptr {
	return s.fd
}

// IsTerminal reports whether the stream is connected to a terminal.
func (s *CommonStream) IsTerminal() bool {
	return s.isTerminal
}

// SetRawMode places the terminal in raw mode.
func (s *CommonStream) SetRawMode() error {
	if s.state != nil {
		return nil
	}
	state, err := term.SetRawTerminal(s.FD())
	if err != nil {
		return err
	}
	s.state = state
	return nil
}

// RestoreTerminal restores the terminal mode saved by SetRawMode.
func (s *CommonStream) RestoreTerminal() error {
	if s.state == nil {
		return nil
	}
	if err := term.RestoreTerminal(s.fd, s.state); err != nil {
		return err
	}
	s.state = nil
	return nil
}

// SetIsTerminal overrides whether the stream is connected to a terminal.
func (s *CommonStream) SetIsTerminal(isTerminal bool) {
	s.isTerminal = isTerminal
}
