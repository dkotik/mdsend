package sqlite

import (
	"net/mail"
	"testing"

	"github.com/dkotik/mdsend"
	"zombiezen.com/go/sqlite"
)

func TestLetterQueries(t *testing.T) {
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

	qq, err := New(conn, "test_letters_")
	if err != nil {
		t.Fatal(err)
	}
	q := qq.(sqliteQueue)
	letterID := "testLetterForQueries"
	content := "test content"
	ctx := t.Context()

	if err = q.CreateLetter(ctx, mdsend.Letter{
		ID:      letterID,
		Content: content,
	}); err != nil {
		t.Fatal(err)
	}

	messageID := "testMessageForMarkingAsSent"
	if err = q.CreateDispatch(ctx, mdsend.Dispatch{
		ID:       messageID,
		LetterID: letterID,
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

	ok, err := q.MarkLetterAsSent(ctx, letterID)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected letter to be marked as yet not sent:", q.DB.Changes())
	}

	if _, err = q.MarkMessageAsSent(ctx, messageID); err != nil {
		t.Fatal(err)
	}
	if q.DB.Changes() == 0 {
		t.Error("expected dispatch to be marked as complete:", q.DB.Changes())
	}
	ok, err = q.MarkLetterAsSent(ctx, letterID)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected letter to be marked as sent:", q.DB.Changes())
	}
}
