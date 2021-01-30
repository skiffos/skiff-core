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
	in := NewInStream(stdIn, false)
	out := NewOutStream(stdOut)
	errOut := NewOutStream(stdErr)
	inStrm, _ := in.(*InStream)
	useTty := inStrm != nil && inStrm.IsTty()

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

	strm := &HijackedIOStreamer{
		Resp: conn,
		Tty:  useTty,
	}
	if in != nil {
		strm.InputStream = in
	}
	if out != nil {
		strm.OutputStream = out
	}
	if errOut != nil {
		strm.ErrorStream = errOut
	}
	return strm.Stream(ctx)
}
