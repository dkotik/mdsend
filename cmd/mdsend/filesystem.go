package main

import (
	"io"
	"os"
	"path/filepath"
)

type unsafeUnconstrainedFileSystem struct{}

func (fs unsafeUnconstrainedFileSystem) OpenFile(path string) (io.ReadCloser, error) {
	// Directly pass unconstrained paths to the OS
	// without sanitizing. Support Windows path.
	return os.Open(filepath.FromSlash(path))
}
