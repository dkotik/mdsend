package sqlite

import (
	q "github.com/dkotik/mdsend/queue"
	"zombiezen.com/go/sqlite"
)

type queue struct {
	DB *sqlite.Conn
}

// New creates an SQLite3 queue at the location.
func New(conn *sqlite.Conn) (q.Queue, error) {
	return queue{
		DB: conn,
	}, nil
}
