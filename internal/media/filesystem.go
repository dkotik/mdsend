package media

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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
	imports map[string]int
	fs      fs.FS
}

func NewCyclicalImportPreventingFileSystem(
	fs fs.FS,
) fs.FS {
	if fs == nil {
		panic("file system is nil")
	}
	return cyclicalImportPreventingFileSystem{
		fs:      fs,
		imports: make(map[string]int),
	}
}

func (fs cyclicalImportPreventingFileSystem) Open(p string) (fs.File, error) {
	count := fs.imports[p]
	if count > 64 {
		return nil, ErrCyclicalImport
	}
	fs.imports[p] = count + 1
	file, err := fs.fs.Open(p)
	if err != nil {
		return nil, err
	}
	return file, nil
}
