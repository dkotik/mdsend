package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/dkotik/mdsend"
)

type mockMailer struct {
	sentLetters []string
	mu          *sync.Mutex
}

func (m *mockMailer) SendMail(ctx context.Context, letter mdsend.Message) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentLetters = append(m.sentLetters, letter.ID)
	return letter.ID, nil
}

func TestSend(t *testing.T) {
	if testing.Short() {
		t.Skip("slow test")
	}
	database := filepath.Join(t.TempDir(), "cmdSendTest.sqlite3")
	t.Cleanup(func() {
		if err := os.Remove(database); err != nil {
			t.Fatal("failed to clean up database file:", err)
		}
	})
	ctx := t.Context()
	if err := application.Run(ctx, []string{
		"mdsend",
		"--destroy",
		"--queue", database,
		"../../examples/1-minimal.md",
		"../../examples/4-themes.md",
	}); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal("unable to send letters:", err)
	}
}
