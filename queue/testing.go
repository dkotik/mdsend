package queue

import (
	"errors"
	"iter"
	"net/mail"
	"testing"
	"time"

	"github.com/dkotik/mdsend"
)

var mockTime = time.Date(2026, 5, 6, 7, 8, 9, 11, time.FixedZone("PST", -8*60*60))

func IteratorIsEmpty[T any](
	iterator iter.Seq2[T, error],
) func(*testing.T) {
	return func(t *testing.T) {
		for _, err := range iterator {
			if err != nil {
				t.Fatal(err)
			}
			t.Fatal("found one item")
		}
	}
}

func TestQueue(q Queue) func(*testing.T) {
	l1 := mdsend.Letter{
		ID: "firstLetter",
		Frontmatter: map[string]any{
			"subject": "first letter",
		},
		Content:   "first letter",
		CreatedAt: mockTime,
		SentAt:    mockTime.Add(time.Hour * 70),
	}

	l2 := mdsend.Letter{
		ID: "secondLetter",
		Frontmatter: map[string]any{
			"subject": "second letter",
		},
		Content:   "second letter",
		CreatedAt: mockTime.Add(time.Hour * 50),
		SentAt:    mockTime.Add(time.Hour * 170),
	}

	messages := []mdsend.Message{
		{
			ID:       "firstMessage",
			LetterID: l1.ID,
			From:     mail.Address{},
			To: mail.Address{
				Name:    "First Last",
				Address: "first.last@example.com",
			},
			Subject: "subject1",
			Text:    "first text",
			HTML:    "first HTML",
		},
		{
			ID:       "secondMessage",
			LetterID: l1.ID,
			From:     mail.Address{},
			To: mail.Address{
				Name:    "Second",
				Address: "second@example.com",
			},
			Subject: "subject2",
			Text:    "second text",
			HTML:    "second HTML",
		},
	}

	attachments := []mdsend.Attachment{
		{
			LetterID: l1.ID,
			Content:  []byte("attachment content1"),
			Hash:     "attachment content1",
		},
		{
			LetterID:  l1.ID,
			ContentID: "<inline@domain.com>",
			Content:   []byte("attachment content2"),
			Hash:      "attachment content2",
		},
	}
	return func(t *testing.T) {
		var (
			ctx = t.Context()
			err error
		)
		if err = q.CreateLetter(ctx, l1); err != nil {
			t.Fatal("unable to create the first letter:", err)
		}

		for _, a := range attachments {
			if err = q.CreateAttachment(ctx, a); err != nil {
				t.Fatal("unable to create attachment:", err)
			}
		}

		for _, d := range messages {
			if err = q.CreateMessage(ctx, d); err != nil {
				t.Fatal("unable to create message:", err)
			}
		}

		defer func() {
			if err = q.DeleteLetter(ctx, l1.ID); err != nil {
				t.Fatal("unable to delete the first letter:", err)
			}
			for l1, err = range q.ListLetters(ctx, Cursor{Batch: 1}) {
				if err != nil {
					t.Fatal("unable to collect a list of letters:", err)
				}
				t.Fatal("There is still a letter left over:", l1)
			}

			t.Run("there are no messages left over", IteratorIsEmpty(q.ListMessages(ctx, ChildCursor{
				ParentID: l1.ID,
				Cursor: Cursor{
					ItemID: "",
					Batch:  5,
				},
			})))

			t.Run("there are no attachments left over", IteratorIsEmpty(q.ListAttachments(ctx, l1.ID)))
		}()

		lcomp, err := q.RetrieveLetter(ctx, l1.ID)
		if err != nil {
			t.Fatal("unable to retrieve first letter:", err)
		}
		if err = l1.AssertEqualityTo(lcomp); err != nil {
			t.Fatal("letters do not match:", err)
		}

		l1attachments := make([]mdsend.Attachment, 0, len(attachments))
		for l1attachment, err := range q.ListAttachments(ctx, l1.ID) {
			if err != nil {
				t.Fatal("unable to list attachments for first letter:", err)
			}
			l1attachments = append(l1attachments, l1attachment)
		}
		if len(l1attachments) != len(attachments) {
			t.Fatal("attachment count mismatch: expected", len(attachments), "got", len(l1attachments))
		}
		for i, a := range l1attachments {
			if err = a.AssertEqualityTo(attachments[i]); err != nil {
				t.Fatal("attachments do not match:", err)
			}
		}

		l1messages := make([]mdsend.Message, 0, 1)
		for l1message, err := range q.ListMessages(ctx, ChildCursor{
			ParentID: l1.ID,
			Cursor: Cursor{
				ItemID: "",
				Batch:  1,
			},
		}) {
			if err != nil {
				t.Fatal("unable to list messages for first letter:", err)
			}
			l1messages = append(l1messages, l1message)
		}
		if len(l1messages) != len(messages) {
			t.Fatal("message count mismatch: expected", len(messages), "got", len(l1messages))
		}
		for i, d := range l1messages {
			d.ID = messages[i].ID // copy the ID from the expected message
			if err = d.AssertEqualityTo(messages[i]); err != nil {
				t.Fatal("messages do not match:", err)
			}
			err := q.MarkMessagesAsQueued(ctx, d.ID)
			if err != nil {
				t.Fatalf("unable to complete message %d: %v", i+1, err)
			}
			ok, err := q.MarkMessageAsSent(ctx, d.ID)
			if err != nil {
				t.Fatalf("unable to complete message %d: %v", i+1, err)
			}
			if !ok {
				t.Fatalf("message %d was not marked as sent", i+1)
			}
		}

		if err = q.CreateLetter(ctx, l2); err != nil {
			t.Fatal("unable to create the second letter:", err)
		}
		defer func() {
			if err = q.DeleteLetter(ctx, l2.ID); err != nil {
				t.Fatal("unable to delete the second letter:", err)
			}

			t.Run("there are no messages left over", IteratorIsEmpty(q.ListMessages(ctx, ChildCursor{
				ParentID: l2.ID,
				Cursor: Cursor{
					ItemID: "",
					Batch:  5,
				},
			})))

			t.Run("there are no attachments left over", IteratorIsEmpty(q.ListAttachments(ctx, l2.ID)))
		}()

		lcomp, err = q.RetrieveLetter(ctx, l2.ID)
		if err != nil {
			t.Fatal("unable to retrieve second letter:", err)
		}
		if err = l2.AssertEqualityTo(lcomp); err != nil {
			t.Fatal("letters do not match:", err)
		}

		// test letter listing
		ok := false
		next, stop := iter.Pull2(q.ListLetters(ctx, Cursor{Batch: 1}))
		if next == nil {
			t.Fatal("no letters found")
		}
		lcomp, err, ok = next()
		if !ok {
			t.Fatal("no letters found, when the first letter was expected")
		}
		if err != nil {
			t.Fatal("unable to retrieve the first letter:", err)
		}
		if err = l1.AssertEqualityTo(lcomp); err != nil {
			t.Fatal("letters do not match:", err)
		}
		lcomp, err, ok = next()
		if !ok {
			t.Fatal("no letters found, when the second letter was expected")
		}
		if err != nil {
			t.Fatal("unable to retrieve the second letter:", err)
		}
		if err = l2.AssertEqualityTo(lcomp); err != nil {
			t.Fatal("letters do not match:", err)
		}
		stop()

		t.Run("queue recognizes duplicates", QueueRecognizesDuplicates(q))

		t.Run("transaction support", func(t *testing.T) {
			qtx, tx, err := q.BeginTransaction(ctx)
			if err != nil {
				t.Fatal("unable to begin transaction:", err)
			}
			qtx, err = qtx.WithTransaction(ctx, tx)
			if err != nil {
				t.Fatal("unable to with transaction:", err)
			}

			txLetterID := "txTestLetter"
			if err = qtx.CreateLetter(ctx, mdsend.Letter{
				ID: txLetterID,
			}); err != nil {
				t.Fatal("unable to create letter:", err)
			}
			err = errors.New("cause transaction to fail")
			tx.Close(&err)

			if _, err = q.RetrieveLetter(ctx, txLetterID); err == nil {
				t.Fatal("letter should not be retrievable after transaction failure")
			}
		})
	}
}

func QueueRecognizesDuplicates(q Queue) func(*testing.T) {
	return func(t *testing.T) {
		ctx := t.Context()
		l1 := mdsend.Letter{
			ID: "duplicationTestLetter",
			Frontmatter: map[string]any{
				"subject": "first letter",
			},
			Content:   "first letter",
			CreatedAt: mockTime.Add(time.Hour * 170),
			SentAt:    mockTime.Add(time.Hour * 270),
		}
		if err := q.CreateLetter(ctx, l1); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := q.DeleteLetter(ctx, l1.ID); err != nil {
				t.Fatal(err)
			}
		}()

		d := mdsend.Message{
			LetterID: "testLetter",
			From:     mail.Address{},
			To: mail.Address{
				Name:    "First Last",
				Address: "first.last@example.com",
			},
			Subject: "",
			Text:    "",
			HTML:    "",
		}
		if err := q.CreateMessage(ctx, d); err != nil {
			t.Fatal(err)
		}
		if err := q.CreateMessage(ctx, d); !errors.Is(err, mdsend.ErrDuplicateMessage) {
			t.Fatalf("expected duplicate message error, got: %v", err)
		}

		a := mdsend.Attachment{
			Hash:        "testAttachmentForDuplicates",
			LetterID:    "testLetter",
			Name:        "test.txt",
			ContentType: "text/plain",
			Content:     []byte("test"),
		}
		if err := q.CreateAttachment(ctx, a); err != nil {
			t.Fatal(err)
		}
		if err := q.CreateAttachment(ctx, a); !errors.Is(err, mdsend.ErrDuplicateAttachment) {
			t.Fatalf("expected duplicate attachment error, got: %v", err)
		}
	}
}
