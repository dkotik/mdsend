package sqlite

import (
	"net/mail"
	"testing"

	"github.com/dkotik/mdsend"
	"zombiezen.com/go/sqlite"
)

func TestDispatchQueries(t *testing.T) {
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
	q := qq.(queue)
	letterID := "testLetter"
	content := "test content"
	ctx := t.Context()

	if err = q.CreateLetter(ctx, mdsend.Letter{
		ID:      letterID,
		Content: content,
	}, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err = q.CreateDispatch(ctx, mdsend.Dispatch{
		LetterID: letterID,
		From:     mail.Address{},
		To:       mail.Address{},
		Subject:  "",
		Text:     "",
		HTML:     "",
	}); err != nil {
		t.Fatal(err)
	}

	if err = q.CreateDispatch(ctx, mdsend.Dispatch{
		LetterID: letterID,
		From:     mail.Address{},
		To:       mail.Address{},
		Subject:  "",
		Text:     "",
		HTML:     "",
	}); err != nil {
		t.Fatal(err)
	}

	dispatches := make([]mdsend.Dispatch, 0, 2)
	for d, err := range q.ListDispatches(ctx, mdsend.ChildCursor{
		ParentID: letterID,
		Cursor: mdsend.Cursor{
			ItemID: "",
			Batch:  1,
		},
	}) {
		if err != nil {
			t.Fatal(err)
		}
		dispatches = append(dispatches, d)
	}
	if len(dispatches) != 2 {
		t.Fatal("expected 2 attachments, got", len(dispatches))
	}
}
