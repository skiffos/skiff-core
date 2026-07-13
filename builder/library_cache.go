package builder

import (
	"os"
	"sync"

	"github.com/paralin/scratchbuild/library"
)

type libraryCache struct {
	// mtx guards all fields
	mtx sync.Mutex
	// refCount is the number of active builders
	refCount int
	// path is the temporary library cache directory
	path string
	// library is the shared resolver
	library *library.LibraryResolver
}

var globalLibraryCache = &libraryCache{}

func (c *libraryCache) get() (*library.LibraryResolver, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.refCount > 0 {
		c.refCount++
		return c.library, nil
	}

	path, err := os.MkdirTemp("", "skiff-core-scratch-")
	if err != nil {
		return nil, err
	}
	libraryResolver, err := library.BuildLibraryResolver(path)
	if err != nil {
		os.RemoveAll(path)
		return nil, err
	}
	c.path = path
	c.library = libraryResolver
	c.refCount = 1
	return libraryResolver, nil
}

func (c *libraryCache) release() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.refCount == 0 {
		return nil
	}
	c.refCount--
	if c.refCount != 0 {
		return nil
	}

	path := c.path
	c.path = ""
	c.library = nil
	return os.RemoveAll(path)
}
