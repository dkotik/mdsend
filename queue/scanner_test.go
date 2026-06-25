package queue_test

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"testing"
	"testing/synctest"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"golang.org/x/sync/errgroup"
	"zombiezen.com/go/sqlite"
)

func TestContinuousScanner(t *testing.T) {
	if testing.Short() {
		t.Skip("scanner reads and writes many records")
	}
	dsn := ":memory:"
	conn, err := sqlite.OpenConn(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	q, err := sqliteQ.New(conn, "")
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	if err = q.CreateLetter(ctx, mdsend.Letter{
		ID: "test-letter-id-1",
	}); err != nil {
		t.Fatal(err)
	}
	for i := range 1000 {
		if err = q.CreateMessage(ctx, mdsend.Message{
			ID: fmt.Sprintf("message_%d", i),
			To: mail.Address{
				Name:    "Recipient",
				Address: fmt.Sprintf("recipient_%d@test.com", i),
			},
			LetterID: "test-letter-id-1",
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err = q.CreateLetter(ctx, mdsend.Letter{
		ID: "test-letter-id-2",
	}); err != nil {
		t.Fatal(err)
	}
	for i := range 1000 {
		if err = q.CreateMessage(ctx, mdsend.Message{
			ID: fmt.Sprintf("message_%d", i+2000),
			To: mail.Address{
				Name:    "Recipient",
				Address: fmt.Sprintf("recipient_%d@test.com", i),
			},
			LetterID: "test-letter-id-2",
		}); err != nil {
			t.Fatal(err)
		}
	}
	// for letter, err := range q.ListLetters(ctx, queue.Cursor{Batch: -2}) {
	// 	t.Log("found letter:", letter.ID, err)
	// }

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		eg, ctx := errgroup.WithContext(ctx)
		found := 0
		queue.NewContinuousScanner(ctx, eg, q, queue.SchedulerFunc(func(ctx context.Context, m []mdsend.Message) error {
			found++
			return nil
		}), queue.ContinuousScannerOptions{
			Frequency: time.Millisecond * 30,
			// BeginWithOlderLetters: true,
		})

		<-time.After(time.Minute * 2)
		if found < 2000 {
			t.Fatal("did not find enough records:", found)
		}

		cancel()
		if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, sqlite.ResultInterrupt.ToError()) {
			t.Fatal("errgroup wait error: ", err)
		}
	})
}
