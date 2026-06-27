package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dkotik/mdsend"
	"golang.org/x/sync/errgroup"
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
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	var err error
	// userDir, err := os.UserHomeDir()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// dsn := "file:" + filepath.Join(userDir, "Downloads", "mdsend.sqlite3?cache=shared&wal=on")
	dsn := fmt.Sprintf("%s?cache=shared&foreign_keys=on&wal=on", filepath.Join(t.TempDir(), "testSend.sqlite3"))
	// dsn := "file:/test/sendCommand?vfs=memdb"
	// dsn := "file:sendTest?mode=memory&cache=shared&foreign_keys=on"
	err = addLetters(ctx, dsn, []mdsend.Letter{
		mdsend.Letter{
			ID: "firstTestLetter",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	b := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(b, &slog.HandlerOptions{
		Level: slog.Level(slog.LevelDebug - 100),
	}))

	mailer := &mockMailer{
		sentLetters: []string{},
		mu:          &sync.Mutex{},
	}
	wg, ctx := errgroup.WithContext(ctx)
	err = send(
		ctx,
		wg,
		dsn,
		time.Second/8,
		// newSemaphoreMailer(6),
		mailer,
		logger,
	)
	if err != nil {
		t.Fatal(err)
	}
	err = wg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		for _, line := range bytes.Split(b.Bytes(), []byte("\n")) {
			t.Log(string(line))
		}
		t.Fatal(err)
	}

	if len(mailer.sentLetters) < 1000 {
		for _, line := range bytes.Split(b.Bytes(), []byte("\n")) {
			t.Log(string(line))
		}
		t.Fatalf("expected 1 sent message, got %d", len(mailer.sentLetters))
	}
}
