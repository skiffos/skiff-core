package setup

import (
	"context"
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
)

type containerExecResult struct {
	stderr string
	err    error
}

type recordingContainerWaiter struct {
	results  []containerExecResult
	commands []string
}

func (w *recordingContainerWaiter) CheckHasContainer(string) bool {
	return true
}

func (w *recordingContainerWaiter) WaitForContainer(context.Context, string, io.Writer) (string, error) {
	return "container-id", nil
}

func (w *recordingContainerWaiter) ExecCmdContainer(
	_ context.Context,
	_ string,
	_ string,
	_ io.Reader,
	_ io.Writer,
	stderr io.Writer,
	command string,
	_ ...string,
) error {
	w.commands = append(w.commands, command)
	result := w.results[len(w.commands)-1]
	if result.stderr != "" && stderr != nil {
		if _, err := io.WriteString(stderr, result.stderr); err != nil {
			return err
		}
	}
	return result.err
}

func TestEnsureContainerUserReturnsUnexpectedLookupError(t *testing.T) {
	expectedError := errors.New("container exec failed")
	waiter := &recordingContainerWaiter{results: []containerExecResult{{
		stderr: "permission denied",
		err:    expectedError,
	}}}
	userSetup := NewUserSetup(
		logrus.NewEntry(logrus.New()),
		&config.ConfigUser{ContainerUser: "core"},
		waiter,
		false,
	)

	err := userSetup.ensureContainerUser(t.Context(), "container-id")
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected lookup error, got %v", err)
	}
	if len(waiter.commands) != 1 || waiter.commands[0] != "id" {
		t.Fatalf("expected one id command, got %v", waiter.commands)
	}
}

func TestEnsureContainerUserReturnsCreationError(t *testing.T) {
	expectedError := errors.New("useradd failed")
	waiter := &recordingContainerWaiter{results: []containerExecResult{
		{stderr: "id: core: no such user\n", err: errors.New("missing user")},
		{err: expectedError},
	}}
	userSetup := NewUserSetup(
		logrus.NewEntry(logrus.New()),
		&config.ConfigUser{ContainerUser: "core"},
		waiter,
		false,
	)

	err := userSetup.ensureContainerUser(t.Context(), "container-id")
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected creation error, got %v", err)
	}
	if len(waiter.commands) != 2 || waiter.commands[0] != "id" || waiter.commands[1] != "useradd" {
		t.Fatalf("expected id then useradd, got %v", waiter.commands)
	}
}
