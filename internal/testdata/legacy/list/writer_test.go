package list

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRenamingWriter(t *testing.T) {
	target := filepath.Join(t.TempDir(), "testing.md")
	w, err := NewWriter(target)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(w, "# Testing\n"); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}

	saved, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(saved, []byte("# Testing\n")) {
		t.Fatal("expected result does not match the test file")
	}
}
