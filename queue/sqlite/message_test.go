package sqlite

import (
	"net/mail"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
	"github.com/oklog/ulid/v2"
	"zombiezen.com/go/sqlite"
)

func TestMessageQueries(t *testing.T) {
	conn, err := sqlite.OpenConn("file::memory:?cache=shared&?_foreign_keys=true")
	if err != nil {
		t.Fatal("unable to open SQLite3 connection:", err)
	}
	defer conn.Close()
	// if err = sqlitex.ExecScript(conn, `
	// PRAGMA foreign_keys = ON;
	// PRAGMA strict = ON;
	// 		`); err != nil {
	// 	t.Fatal("unable to set foreign keys:", err)
	// }

	qq, err := New(conn, "test_attachments_")
	if err != nil {
		t.Fatal(err)
	}
	q := qq.(sqliteQueue)
	letterID := "testLetter"
	content := "test content"
	ctx := t.Context()

	if err = q.CreateLetter(ctx, mdsend.Letter{
		ID:      letterID,
		Content: content,
	}); err != nil {
		t.Fatal(err)
	}

	if err = q.CreateMessage(ctx, mdsend.Message{
		ID:       ulid.Make().String(),
		LetterID: letterID,
		SeedKey:  letterID + "1",
		From:     mail.Address{},
		To: mail.Address{
			Name:    "First Last",
			Address: "first.last@example.com",
		},
		Subject: "",
		Text:    "",
		HTML:    "",
	}); err != nil {
		t.Fatal(err)
	}

	if err = q.CreateMessage(ctx, mdsend.Message{
		ID:       ulid.Make().String(),
		LetterID: letterID,
		SeedKey:  letterID + "1",
		From:     mail.Address{},
		To: mail.Address{
			Name:    "Second",
			Address: "second@example.com",
		},
		Subject: "",
		Text:    "",
		HTML:    "",
	}); err != nil {
		t.Fatal(err)
	}

	messages := make([]mdsend.Message, 0, 2)
	for d, err := range q.ListMessages(ctx, queue.ChildCursor{
		ParentID: letterID,
		Cursor: queue.Cursor{
			ItemID: "",
			Batch:  1,
		},
	}) {
		if err != nil {
			t.Fatal(err)
		}
		messages = append(messages, d)
	}
	if len(messages) != 2 {
		t.Fatal("expected 2 messages, got", len(messages))
	}
}
