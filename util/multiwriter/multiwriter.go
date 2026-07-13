package multiwriter

import (
	"io"
	"sync"
)

// MultiWriter writes to a dynamically managed set of writers.
type MultiWriter struct {
	// mtx guards writers and serializes writes
	mtx sync.Mutex
	// writers contains the current output set
	writers []io.Writer
}

// AddWriter adds a non-nil output writer.
func (w *MultiWriter) AddWriter(writer io.Writer) {
	if writer == nil {
		return
	}
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.writers = append(w.writers, writer)
}

// RemoveWriter removes the first matching output writer.
func (w *MultiWriter) RemoveWriter(writer io.Writer) {
	if writer == nil {
		return
	}
	w.mtx.Lock()
	defer w.mtx.Unlock()
	for index, candidate := range w.writers {
		if candidate == writer {
			w.writers[index] = w.writers[len(w.writers)-1]
			w.writers[len(w.writers)-1] = nil
			w.writers = w.writers[:len(w.writers)-1]
			return
		}
	}
}

// Write writes the complete buffer to each current writer.
func (w *MultiWriter) Write(p []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	for _, writer := range w.writers {
		written, err := writer.Write(p)
		if err != nil {
			return written, err
		}
		if written != len(p) {
			return written, io.ErrShortWrite
		}
	}
	return len(p), nil
}

// _ is a type assertion.
var _ io.Writer = ((*MultiWriter)(nil))
