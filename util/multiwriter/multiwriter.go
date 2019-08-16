package multiwriter

import (
	"io"
	"sync"
)

// MultiWriter manages concurrent loggers
type MultiWriter struct {
	mtx     sync.Mutex
	writers []io.Writer
}

func (w *MultiWriter) AddWriter(wr io.Writer) {
	if wr == nil {
		return
	}
	w.mtx.Lock()
	w.writers = append(w.writers, wr)
	w.mtx.Unlock()
}

func (w *MultiWriter) RmWriter(wr io.Writer) {
	if wr == nil {
		return
	}
	w.mtx.Lock()
	for i, wri := range w.writers {
		if wri == wr {
			w.writers[i] = w.writers[len(w.writers)-1]
			w.writers[len(w.writers)-1] = nil
			w.writers = w.writers[:len(w.writers)-1]
			break
		}
	}
	w.mtx.Unlock()
}

func (w *MultiWriter) Write(p []byte) (n int, err error) {
	w.mtx.Lock()
	for _, wr := range w.writers {
		wr.Write(p)
	}
	w.mtx.Unlock()
	return len(p), nil
}

// _ is a type assertion
var _ io.Writer = ((*MultiWriter)(nil))
