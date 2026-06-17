package main

import (
	"errors"

	"github.com/adrg/xdg"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
	repository "github.com/dkotik/mdsend/queue/sqlite"
	"github.com/urfave/cli/v3"
	"zombiezen.com/go/sqlite"
)

var flagDatabase = &cli.StringFlag{
	Name:    `database`,
	Usage:   `Path to the queue database file or data source name.`,
	Aliases: []string{`db`},
	Sources: cli.ValueSourceChain{
		Chain: []cli.ValueSource{
			cli.EnvVar("MDSEND_DATABASE"),
			xdgDataFile("queue.sqlite3"),
		},
	},
	Value: "mdsend_queue.sqlite3",
}

func newQueue(l mdsend.Letter) (queue.Queue, func() error, error) {
	queue, err := l.GetQueue()
	if err != nil {
		if !errors.Is(err, mdsend.ErrNoQueue) {
			return nil, nil, err
		}
		queue, err = xdg.DataFile("mdsend/queue.sqlite3")
		if err != nil {
			return nil, nil, err
		}
	}
	conn, err := sqlite.OpenConn(queue, sqlite.OpenReadWrite)
	if err != nil {
		return nil, nil, err
	}
	q, err := repository.New(conn, "")
	if err != nil {
		return nil, nil, err
	}
	return q, conn.Close, nil
}
