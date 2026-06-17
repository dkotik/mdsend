package test

import (
	"errors"
	"net/mail"
	"testing"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal"
	"github.com/dkotik/mdsend/queue"
)

func QueueRecognizesDuplicates(q queue.Queue) func(*testing.T) {
	return func(t *testing.T) {
		ctx := t.Context()
		l1 := mdsend.Letter{
			ID: "duplicationTestLetter",
			Frontmatter: map[string]any{
				"subject": "first letter",
			},
			Content:   "first letter",
			CreatedAt: internal.MockTime.Add(time.Hour * 170),
			SentAt:    internal.MockTime.Add(time.Hour * 270),
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
