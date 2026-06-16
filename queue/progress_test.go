package queue_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"testing"
	"testing/synctest"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"golang.org/x/sync/errgroup"
	"zombiezen.com/go/sqlite"
)

func TestProgressTracker(t *testing.T) {
	if testing.Short() {
		t.Skip("progress tracker reads and writes many records")
	}
	dsn := "file:shared_memory_db?mode=memory&cache=shared"
	conn, err := sqlite.OpenConn(dsn)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = conn.Close(); err != nil {
			panic(err)
		}
	}()
	q, err := sqliteQ.New(conn, "")
	if err != nil {
		panic(err)
	}
	conn2, err := sqlite.OpenConn(dsn)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = conn2.Close(); err != nil {
			panic(err)
		}
	}()
	q2, err := sqliteQ.New(conn2, "")
	if err != nil {
		panic(err)
	}

	const testLetterID = "test-letter-id"
	ctx := t.Context()
	if err = q.CreateLetter(ctx, mdsend.Letter{
		ID: testLetterID,
	}); err != nil {
		panic(err)
	}

	for i := range 1000 {
		if err = q.CreateDispatch(ctx, mdsend.Dispatch{
			ID: fmt.Sprintf("message_%d", i),
			To: mail.Address{
				Name:    "Recipient",
				Address: fmt.Sprintf("recipient_%d@test.com", i),
			},
			LetterID: testLetterID,
		}); err != nil {
			panic(err)
		}
	}

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		eg, ctx := errgroup.WithContext(ctx)
		frequency := time.Second
		scanner, queuedMessageIDs := queue.NewScanner(
			frequency,
			queue.Cursor{},
			queue.ChildCursor{},
		)
		scanner.JoinErrorGroup(ctx, eg, q)

		tracker, handler, progress := queue.NewProgressTracker(
			queuedMessageIDs,
		)
		tracker.JoinErrorGroup(ctx, eg, q2)
		b := &bytes.Buffer{}
		encoder := json.NewEncoder(b)
		confirmOne := func(id int) (err error) {
			b.Reset()
			if err = encoder.Encode(queue.Confirmation{
				LetterID:       testLetterID,
				MessageID:      fmt.Sprintf("message_%d", id),
				ConfirmationID: fmt.Sprintf("confirmation_%d", id),
				// SentAt:         time.Now(),
			}); err != nil {
				return err
			}
			m := message.NewMessage(
				fmt.Sprintf("uuid_%d", id),
				b.Bytes(),
			)
			m.SetContext(ctx)
			if err = handler(m); err != nil {
				return err
			}
			return nil
		}

		for i := range 333 {
			if err = confirmOne(i); err != nil {
				t.Fatal(err)
			}
		}
		<-time.After(frequency * 2)
		<-progress
		p := <-progress
		if p.Sent != 333 || p.Total < 333 {
			t.Error("progress tracker: ", p)
		}

		for i := range 335 {
			if err = confirmOne(335 + i); err != nil {
				t.Fatal(err)
			}
		}
		<-time.After(frequency * 3)
		<-progress
		p = <-progress
		if p.Sent != 668 || p.Total < 668 {
			t.Error("progress tracker: ", p)
		}

		for i := range 334 {
			if err = confirmOne(668 + i); err != nil {
				t.Fatal(err)
			}
		}
		<-time.After(frequency * 3)
		<-progress
		p = <-progress
		if p.Sent != 1000 || p.Total < 900 {
			t.Error("progress tracker: ", p)
		}

		cancel()
		if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal("errgroup wait error: ", err)
		}
	})
}
