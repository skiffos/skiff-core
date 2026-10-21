package execcmd

import (
	"context"
	"testing"

	"github.com/moby/moby/client"
	"github.com/pkg/errors"
)

type staticExecInspector struct {
	result client.ExecInspectResult
	err    error
}

func (s staticExecInspector) ExecInspect(context.Context, string, client.ExecInspectOptions) (client.ExecInspectResult, error) {
	return s.result, s.err
}

func TestInspectExecExitPreservesStatus(t *testing.T) {
	testCases := []struct {
		name     string
		exitCode int
	}{
		{name: "success", exitCode: 0},
		{name: "false", exitCode: 1},
		{name: "explicit status", exitCode: 42},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := InspectExecExit(context.Background(), staticExecInspector{
				result: client.ExecInspectResult{ExitCode: testCase.exitCode},
			}, "exec-id")
			if testCase.exitCode == 0 {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				return
			}

			var exitErr *CommandExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected CommandExitError, got %T: %v", err, err)
			}
			if exitErr.ExitCode() != testCase.exitCode {
				t.Fatalf("expected exit code %d, got %d", testCase.exitCode, exitErr.ExitCode())
			}
		})
	}
}

func TestInspectExecExitRejectsRunningProcess(t *testing.T) {
	err := InspectExecExit(context.Background(), staticExecInspector{
		result: client.ExecInspectResult{Running: true},
	}, "exec-id")
	if err == nil {
		t.Fatal("expected a running-process error")
	}
}
