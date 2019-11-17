package builder

import "io"

// nopCloser wraps readers without a Close()
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }
