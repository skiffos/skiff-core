package builder

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/paralin/scratchbuild/library"
)

// libraryCache makes sure we clone the library just once.
type libraryCache struct {
	mtx      sync.Mutex
	refCount int
	path     string
	lib      *library.LibraryResolver
}

var globalLibraryCache = &libraryCache{}

// GetLibrary gets the library instance.
func (lr *libraryCache) GetLibrary() (*library.LibraryResolver, error) {
	lr.mtx.Lock()
	defer lr.mtx.Unlock()

	if lr.refCount == 0 {
		p, err := ioutil.TempDir("", "skiff-core-scratch-")
		if err != nil {
			return nil, err
		}
		lr.path = p
	} else {
		lr.refCount++
		return lr.lib, nil
	}

	lib, err := library.BuildLibraryResolver(lr.path)
	if err != nil {
		return nil, err
	}
	lr.lib = lib
	lr.refCount++
	return lib, nil
}

// Release decrements the refCount
func (lr *libraryCache) Release() {
	lr.mtx.Lock()
	defer lr.mtx.Unlock()

	if lr.refCount == 0 {
		return
	}

	lr.refCount--
	if lr.refCount == 0 {
		os.RemoveAll(lr.path)
		lr.path = ""
	}
}
