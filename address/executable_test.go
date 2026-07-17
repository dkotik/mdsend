package address

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestTakingRecordsFromExecutable(t *testing.T) {
	path, err := exec.LookPath("sh")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			t.Skip("this is not a Unix machine", path)
		}
		t.Fatal("unable to access external shell:", err)
	}

	fs := os.DirFS(".")
	for entry, err := range eachEntryFromExecutable(
		t.Context(),
		"./testdata/executable.yaml.sh",
		fs,
	) {
		if err != nil {
			t.Fatal("failed to read entry:", err)
		}
		entry, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("wrong type: %T", entry)
		}
		_, err := New(entry)
		if err != nil {
			t.Fatal("invalid contact entry:", err)
		}
	}
	// t.Fatal("check")
}
