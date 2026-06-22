package main

import (
	"context"
	"errors"

	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"github.com/urfave/cli/v3"
	"zombiezen.com/go/sqlite"
)

func cmdValidate(ctx context.Context, c *cli.Command) error {
	conn, err := sqlite.OpenConn(
		":memory:?foreign_keys=true",
		sqlite.OpenCreate, sqlite.OpenReadWrite,
	)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, conn.Close())
	}()
	queue, err := sqliteQ.New(conn, "")
	if err != nil {
		return err
	}
	queue, tx, err := queue.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Close(&err)

	if c.IsSet(flagDatabase.Name) {
		// validate letters and messages in the queue
		return nil
	}
	return nil
}
