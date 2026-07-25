package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCommand(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip("slow test")
	// }
	database := filepath.Join(t.TempDir(), "cmdValidateTest.sqlite3")
	t.Cleanup(func() {
		// TODO: remove ErrNotExist condition after real database is used
		if err := os.Remove(database); err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Fatal("failed to clean up database file:", err)
		}
	})
	ctx := t.Context()
	if err := application.Run(ctx, []string{
		"mdsend",
		"validate",
		"--queue", database,
		"../../examples/1-minimal.md",
		"../../examples/2-attachments.md",
		"../../examples/3-scheduling.md",
		"../../examples/4-themes.md",
		"../../examples/5-templates.md",
		"../../examples/6-list.md",
		"../../examples/7-extending.md",
	}); err != nil {
		t.Fatal("unable to queue letters to database:", err)
	}
}
