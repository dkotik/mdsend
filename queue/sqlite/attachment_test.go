package sqlite

import (
	"testing"

	"github.com/dkotik/mdsend"
	"zombiezen.com/go/sqlite"
)

func TestAttachmentQueries(t *testing.T) {
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
	content := []byte("test content")
	ctx := t.Context()

	if err = q.CreateLetter(ctx, mdsend.Letter{
		ID: letterID,
	}); err != nil {
		t.Fatal(err)
	}

	if err = q.CreateAttachment(ctx, mdsend.Attachment{
		LetterID:    letterID,
		Name:        "first",
		Source:      "",
		Hash:        "first",
		ContentType: "test",
		Content:     content,
	}); err != nil {
		t.Fatal(err)
	}

	if err = q.CreateAttachment(ctx, mdsend.Attachment{
		LetterID:    letterID,
		Name:        "second",
		Source:      "",
		Hash:        "second",
		ContentType: "test",
		Content:     content,
	}); err != nil {
		t.Fatal(err)
	}

	attachments := make([]mdsend.Attachment, 0, 2)
	for a, err := range q.ListAttachments(ctx, letterID) {
		if err != nil {
			t.Fatal(err)
		}
		if a.ContentType != "test" {
			t.Fatal("expected content type 'test', got", a.ContentType)
		}
		if string(a.Content) != "test content" {
			t.Fatal("expected content 'test content', got", string(a.Content))
		}
		attachments = append(attachments, a)
	}
	if len(attachments) != 2 {
		t.Fatal("expected 2 attachments, got", len(attachments))
	}

	if err = q.DeleteLetter(ctx, letterID); err != nil {
		t.Fatal(err)
	}
	for a, err := range q.ListAttachments(ctx, letterID) {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("letter attachment found, when it should have been deleted by CASCADE:", a.Name)
	}
}
