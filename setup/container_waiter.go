package setup

import (
	"context"
	"io"
)

// ContainerWaiter waits for container setup and executes container commands.
type ContainerWaiter interface {
	CheckHasContainer(string) bool
	WaitForContainer(context.Context, string, io.Writer) (string, error)
	ExecCmdContainer(
		context.Context,
		string,
		string,
		io.Reader,
		io.Writer,
		io.Writer,
		string,
		...string,
	) error
}
