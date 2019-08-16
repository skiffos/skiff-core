package setup

import (
	"io"
)

// ContainerWaiter waits for a container to be ready.
type ContainerWaiter interface {
	CheckHasContainer(name string) bool
	WaitForContainer(name string, logOut io.Writer) (string, error)
}
