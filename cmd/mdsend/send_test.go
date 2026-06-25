package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
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

func (m mockMailer) SendMail(ctx context.Context, letter mdsend.Message) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentLetters = append(m.sentLetters, letter.ID)
	return letter.ID, nil
}

func TestSend(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	dsn := "file:testSendCommand?mode=memory&cache=shared"
	// dsn := "testdata/testSendCommand?cache=shared"
	err := addLetters(ctx, dsn, []mdsend.Letter{
		mdsend.Letter{},
	})
	if err != nil {
		t.Fatal(err)
	}

	b := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(b, &slog.HandlerOptions{
		Level: slog.Level(slog.LevelDebug - 100),
	}))

	mailer := mockMailer{
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

	// if len(mailer.sentLetters) != 1 || true {
	// 	t.Fatalf("expected 1 sent message, got %d", len(mailer.sentLetters))
	// }
}
