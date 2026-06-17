package list

import (
	"io"
	"os"
	"path/filepath"
)

type renamingWriter struct {
	realPath      string
	temporaryPath string
	temporary     io.WriteCloser
}

// NewWriter creates a resilient file writer. It is resistant
// to incomplete writes due to program crashing or disk filling up
// by using a temporary file, which is renamed to target file
// when the writer is closed successfully.
func NewWriter(p string) (io.WriteCloser, error) {
	f, err := os.CreateTemp(filepath.Dir(p), filepath.Base(p)+".tmp")
	if err != nil {
		return nil, err
	}
	return &renamingWriter{
		realPath:      p,
		temporaryPath: f.Name(),
		temporary:     f,
	}, nil
}

func (w *renamingWriter) Write(b []byte) (int, error) {
	return w.temporary.Write(b)
}

func (w *renamingWriter) Close() (err error) {
	if err = w.temporary.Close(); err != nil {
		return err
	}
	// https://github.com/crawshaw/jsonfile/blob/main/jsonfile.go
	return os.Rename(w.temporaryPath, w.realPath)
}
