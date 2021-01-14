package execcmd

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
)

// ExecCmdContainer executes a command in a Docker container.
func ExecCmdContainer(
	ctx context.Context,
	dockerClient *dockerclient.Client,
	containerID, userID string,
	stdIn io.Reader, stdOut, stdErr io.Writer,
	cmd string, args ...string,
) error {
	var in *InStream
	if stdIn != nil {
		in = NewInStream(stdIn, false)
	}
	var out *OutStream
	if stdOut != nil {
		out = NewOutStream(stdOut)
	}
	var errOut *OutStream
	if stdErr != nil {
		stdErr = NewOutStream(stdErr)
	}
	useTty := in.IsTty()

	cmds := append([]string{cmd}, args...)
	execCreate, err := dockerClient.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Tty:  useTty,
		User: userID,
		Cmd:  cmds,
		// Env: ...,

		AttachStdin:  stdIn != nil,
		AttachStdout: stdOut != nil,
		AttachStderr: stdErr != nil,
	})
	if err != nil {
		return err
	}

	conn, err := dockerClient.ContainerExecAttach(ctx, execCreate.ID, types.ExecStartCheck{
		Tty: useTty,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	return (&HijackedIOStreamer{
		InputStream:  in,
		OutputStream: out,
		ErrorStream:  errOut,
		Resp:         conn,
		Tty:          useTty,
	}).Stream(ctx)
}
