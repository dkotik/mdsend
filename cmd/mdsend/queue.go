package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dkotik/mdsend"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
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

func cmdQueueAdd(ctx context.Context, c *cli.Command) (err error) {
	if c.Args().Len() == 0 {
		return errors.New(`no letters selected to add`)
	}
	letters := make([]mdsend.Letter, 0, c.Args().Len())
	for _, arg := range c.Args().Slice() {
		data, err := os.ReadFile(arg)
		if err != nil {
			return fmt.Errorf(`unable to read file %q: %w`, arg, err)
		}
		letter, err := mdsend.NewLetter(data)
		if err != nil {
			return fmt.Errorf(`unable to parse letter from file %q: %w`, arg, err)
		}
		// TODO: expend letter!
		letters = append(letters, letter)
	}
	p := c.String(flagDatabase.Name)
	if !c.IsSet(flagDatabase.Name) {
		alternativeQueueFile := letters[0].GetQueue()
		if alternativeQueueFile != "" {
			p = alternativeQueueFile
		}
		for _, letter := range letters[1:] {
			if letter.GetQueue() != "" && letter.GetQueue() != p {
				return fmt.Errorf(`atomic operations require all letters to have the same queue: %q vs %q`, letter.GetQueue(), p)
			}
		}
	}

	conn, err := sqlite.OpenConn(
		p,
		sqlite.OpenCreate, sqlite.OpenReadWrite,
	)
	if err != nil {
		return fmt.Errorf("database %q inaccessible: %w", p, err)
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

	if err = queue.CreateLetter(ctx, mdsend.Letter{}); err != nil {
		return err
	}
	for range 3 {
		err = queue.CreateAttachment(ctx, mdsend.Attachment{})
		if err != nil {
			return err
		}
	}
	for range 3 {
		err = queue.CreateMessage(ctx, mdsend.Message{})
		if err != nil {
			return err
		}
	}
	return
}
