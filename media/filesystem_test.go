package media

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCyclicalImports(t *testing.T) {
	fs := NewCyclicalImportPreventingFileSystem(os.DirFS("testdata"))
	var load func(string) error
	load = func(name string) (err error) {
		file, err := fs.Open(name)
		if err != nil {
			return err
		}
		defer func() {
			err = errors.Join(err, file.Close())
		}()
		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		err = load(strings.TrimSpace(string(data)))
		return err
	}

	err := load("a.txt")
	if err != nil && !errors.Is(err, ErrCyclicalImport) {
		t.Fatal(err)
	}
}
