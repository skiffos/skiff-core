package multiwriter

import (
	"io"
	"testing"

	"github.com/pkg/errors"
)

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	return len(p) - 1, nil
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestMultiWriterReportsShortWrites(t *testing.T) {
	writer := &MultiWriter{}
	writer.AddWriter(shortWriter{})

	written, err := writer.Write([]byte("payload"))
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected short-write error, got %v", err)
	}
	if written != len("payload")-1 {
		t.Fatalf("expected partial write count, got %d", written)
	}
}

func TestMultiWriterPropagatesWriterErrors(t *testing.T) {
	expectedError := errors.New("write failed")
	writer := &MultiWriter{}
	writer.AddWriter(failingWriter{err: expectedError})

	written, err := writer.Write([]byte("payload"))
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected writer error, got %v", err)
	}
	if written != 0 {
		t.Fatalf("expected zero bytes written, got %d", written)
	}
}
