package media

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

var ErrCyclicalImport = errors.New("cyclical import of dependencies")

type unsafeUnconstrainedFileSystem struct{}

func NewUnsafeUnconstrainedFileSystem() fs.FS {
	return unsafeUnconstrainedFileSystem{}
}

func (fs unsafeUnconstrainedFileSystem) Open(path string) (fs.File, error) {
	// Directly pass unconstrained paths to the OS
	// without sanitizing. Support Windows path.
	return os.Open(filepath.FromSlash(path))
}

type cyclicalImportPreventingFileSystem struct {
	fs fs.FS

	mu      *sync.Mutex
	imports map[string]int
}

func NewCyclicalImportPreventingFileSystem(
	fs fs.FS,
) fs.FS {
	if fs == nil {
		panic("file system is nil")
	}
	return cyclicalImportPreventingFileSystem{
		fs:      fs,
		mu:      &sync.Mutex{},
		imports: make(map[string]int),
	}
}

func (fs cyclicalImportPreventingFileSystem) Open(p string) (fs.File, error) {
	fs.mu.Lock()
	count := fs.imports[p]
	if count > 64 {
		return nil, ErrCyclicalImport
	}
	fs.imports[p] = count + 1
	fs.mu.Unlock()

	file, err := fs.fs.Open(p)
	if err != nil {
		return nil, err
	}
	return file, nil
}
