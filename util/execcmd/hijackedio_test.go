package execcmd

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sirupsen/logrus"
)

func TestHijackedIOStreamerCopiesStderrWithoutStdout(t *testing.T) {
	var encoded bytes.Buffer
	if _, err := stdcopy.NewStdWriter(&encoded, stdcopy.Stderr).Write([]byte("failure")); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	streamer := &HijackedIOStreamer{
		ErrorStream: &stderr,
		Resp: types.HijackedResponse{
			Reader: bufio.NewReader(&encoded),
		},
		le: logrus.NewEntry(logrus.New()),
	}
	if err := <-streamer.beginOutputStream(); err != nil {
		t.Fatal(err)
	}
	if stderr.String() != "failure" {
		t.Fatalf("expected stderr payload, got %q", stderr.String())
	}
}

func TestTTYHijackedIOStreamerCopiesCombinedOutputToStderrFallback(t *testing.T) {
	var stderr bytes.Buffer
	streamer := &HijackedIOStreamer{
		ErrorStream: &stderr,
		Resp: types.HijackedResponse{
			Reader: bufio.NewReader(bytes.NewBufferString("combined")),
		},
		TTY: true,
		le:  logrus.NewEntry(logrus.New()),
	}
	if err := <-streamer.beginOutputStream(); err != nil {
		t.Fatal(err)
	}
	if stderr.String() != "combined" {
		t.Fatalf("expected combined payload, got %q", stderr.String())
	}
}
