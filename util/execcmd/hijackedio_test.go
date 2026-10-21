package execcmd

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
)

func TestHijackedIOStreamerCopiesStderrWithoutStdout(t *testing.T) {
	// A multiplexed frame: the stream, three zero bytes, and the big-endian
	// length of the payload that follows.
	payload := "failure"
	encoded := bytes.NewBuffer([]byte{byte(stdcopy.Stderr), 0, 0, 0, 0, 0, 0, byte(len(payload))})
	encoded.WriteString(payload)

	var stderr bytes.Buffer
	streamer := &HijackedIOStreamer{
		ErrorStream: &stderr,
		Resp: client.HijackedResponse{
			Reader: bufio.NewReader(encoded),
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
		Resp: client.HijackedResponse{
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
