package queue

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
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"golang.org/x/sync/errgroup"
	"zombiezen.com/go/sqlite"
)

func TestProgressTracker(t *testing.T) {
	if testing.Short() {
		t.Skip("progress tracker reads and writes many records")
	}
	conn, err := sqlite.OpenConn(":memory:")
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
		frequency := time.Second
		ctx, cancel := context.WithCancel(t.Context())
		eg, ctx := errgroup.WithContext(ctx)
		handler, progress := NewProgressTracker(
			ctx,
			q,
			frequency,
			5,
			eg,
		)
		b := &bytes.Buffer{}
		encoder := json.NewEncoder(b)
		confirmOne := func(id int) (err error) {
			b.Reset()
			if err = encoder.Encode(Confirmation{
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
		<-time.After(frequency * 20)
		<-progress
		p := <-progress
		if p.Sent != 333 || p.Total < 333 {
			t.Error("progress tracker: ", p)
		}
		<-time.After(frequency * 20)
		// <-progress
		p = <-progress
		if p.Sent != 333 || p.Total < 333 {
			t.Error("progress tracker: ", p)
		}
		<-time.After(frequency * 20)
		// <-progress
		// p = <-progress
		// if p.Sent != 333 || p.Total < 333 {
		// 	t.Error("progress tracker: ", p)
		// }

		cancel()
		if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal("errgroup wait error: ", err)
		}
	})
}
