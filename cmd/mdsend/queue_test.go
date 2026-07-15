package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dkotik/mdsend/queue"
	"github.com/dkotik/mdsend/queue/sqlite"
)

func TestQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("slow test")
	}
	database := filepath.Join(t.TempDir(), "cmdQueueTest.sqlite3")
	t.Cleanup(func() {
		if err := os.Remove(database); err != nil {
			t.Fatal("failed to clean up database file:", err)
		}
	})
	ctx := t.Context()
	if err := application.Run(ctx, []string{
		"mdsend",
		"queue", "add",
		"--queue", database,
		"../../examples/1-minimal.md",
		"../../examples/5-list.md",
		"../../examples/6-extending.md",
	}); err != nil {
		t.Fatal("unable to queue letters to database:", err)
	}

	conn, err := newQueueConnection(database)
	if err != nil {
		t.Fatal("cannot check database:", err)
	}
	q, err := sqlite.New(conn, "")
	if err != nil {
		t.Fatal("cannot mount queue:", err)
	}

	expectLetters := 3
	foundLetters := make([]string, 0, expectLetters)
	for letter, err := range q.ListLetters(ctx, queue.Cursor{Batch: int64(expectLetters) + 1}) {
		if err != nil {
			t.Fatal("cannot query letters:", err)
		}
		foundLetters = append(foundLetters, letter.ID)
	}
	if len(foundLetters) != expectLetters {
		t.Fatal("unexpected number of letters queued:", foundLetters, "vs", expectLetters)
	}

	expectMessages := 23
	// expectMessages := 45
	foundMessages := 0
	for _, letterID := range foundLetters {
		for _, err = range q.ListMessages(ctx, queue.ChildCursor{
			ParentID: letterID,
			Cursor: queue.Cursor{
				Batch: 100,
			},
		}) {
			if err != nil {
				t.Fatal("unable to query messages:", err)
			}
			foundMessages++
		}
	}

	if foundMessages != expectMessages {
		t.Fatal("unexpected number of messages queued:", foundMessages, "vs", expectMessages)
	}
}
