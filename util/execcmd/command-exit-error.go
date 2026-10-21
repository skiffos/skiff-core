package execcmd

import (
	"context"
	"strconv"

	"github.com/moby/moby/client"
	"github.com/pkg/errors"
)

// ExecInspector inspects the state of a Docker exec process.
type ExecInspector interface {
	// ExecInspect returns the state of the exec process execID.
	ExecInspect(context.Context, string, client.ExecInspectOptions) (client.ExecInspectResult, error)
}

// CommandExitError reports a non-zero Docker exec exit status.
type CommandExitError struct {
	exitCode int
}

// NewCommandExitError constructs an error for a non-zero Docker exec exit status.
func NewCommandExitError(exitCode int) *CommandExitError {
	return &CommandExitError{exitCode: exitCode}
}

// Error returns the command exit error message.
func (e *CommandExitError) Error() string {
	return "command exited with status " + strconv.Itoa(e.exitCode)
}

// ExitCode returns the Docker exec exit status.
func (e *CommandExitError) ExitCode() int {
	return e.exitCode
}

// InspectExecExit returns an error when the Docker exec process did not exit successfully.
func InspectExecExit(ctx context.Context, inspector ExecInspector, execID string) error {
	// A closed stream should mean the process exited; check that it did.
	result, err := inspector.ExecInspect(ctx, execID, client.ExecInspectOptions{})
	if err != nil {
		return errors.Wrap(err, "inspect Docker exec")
	}
	if result.Running {
		return errors.New("docker exec stream closed while command is still running")
	}
	if result.ExitCode != 0 {
		return NewCommandExitError(result.ExitCode)
	}
	return nil
}
