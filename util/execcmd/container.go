package execcmd

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

// ExecCmdContainer executes a command in a Docker container and returns its exit status.
func ExecCmdContainer(
	ctx context.Context,
	le *logrus.Entry,
	dockerClient client.APIClient,
	containerID string,
	userID string,
	stdIn io.Reader,
	stdOut io.Writer,
	stdErr io.Writer,
	cmd string,
	args ...string,
) error {
	in := NewInStream(stdIn, false)
	out := NewOutStream(le, stdOut)
	errOut := NewOutStream(le, stdErr)
	inStream, _ := in.(*InStream)
	useTTY := inStream != nil && inStream.IsTTY()

	command := append([]string{cmd}, args...)
	execCreate, err := dockerClient.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Tty:  useTTY,
		User: userID,
		Cmd:  command,

		AttachStdin:  stdIn != nil,
		AttachStdout: stdOut != nil,
		AttachStderr: stdErr != nil,
	})
	if err != nil {
		return err
	}

	conn, err := dockerClient.ContainerExecAttach(ctx, execCreate.ID, container.ExecAttachOptions{
		Tty: useTTY,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	streamer := NewHijackedIOStreamer(le, conn, useTTY)
	if in != nil {
		streamer.InputStream = in
	}
	if out != nil {
		streamer.OutputStream = out
	}
	if errOut != nil {
		streamer.ErrorStream = errOut
	}
	if err := streamer.Stream(ctx); err != nil {
		return err
	}
	return InspectExecExit(ctx, dockerClient, execCreate.ID)
}
