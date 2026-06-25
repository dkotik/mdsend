package sqlite

import (
	"testing"

	"github.com/dkotik/mdsend/queue"
	"zombiezen.com/go/sqlite"
)

func TestCompliance(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:")
	if err != nil {
		t.Fatal("unable to open SQLite3 connection:", err)
	}

	q, err := New(conn, "")
	if err != nil {
		t.Fatal("unable to create queue:", err)
	}
	queue.TestQueue(q)(t)
}
