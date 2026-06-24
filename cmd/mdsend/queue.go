package main

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/media"
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
	Value: "mdsend_queue.sqlite3?cache=shared&foreign_keys=on",
	Validator: func(p string) error {
		if strings.TrimSpace(p) == "" {
			return errors.New(`database path is empty`)
		}
		return nil
	},
	Action: func(ctx context.Context, c *cli.Command, p string) error {
		p, params, _ := strings.Cut(p, "?")
		paramValues := strings.Split(params, "&")
		if !slices.Contains(paramValues, `cache=shared`) {
			paramValues = append(paramValues, `cache=shared`)
		}
		if !slices.ContainsFunc(paramValues, func(v string) bool {
			return strings.HasPrefix(strings.TrimSpace(v), `foreign_keys=`)
		}) {
			paramValues = append(paramValues, `foreign_keys=on`)
		}
		c.Set(`database`, fmt.Sprintf("%s?%s", p, strings.Join(paramValues, "&")))
		// connectionDSN := "file:ephemeral?mode=memory&cache=shared"
		return nil
	},
}

func cmdQueueAdd(ctx context.Context, c *cli.Command) (err error) {
	if c.Args().Len() == 0 {
		return errors.New(`no letters selected to add`)
	}
	fs := media.NewUnsafeUnconstrainedFileSystem()
	letters := make([]mdsend.Letter, 0, c.Args().Len())
	for _, arg := range c.Args().Slice() {
		letter, err := mdsend.NewLetterFromFile(ctx, fs, arg)
		if err != nil {
			return fmt.Errorf(`unable to parse letter from file %q: %w`, arg, err)
		}
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
	return addLetters(ctx, p, letters)
}

func addLetters(ctx context.Context, conntectionDSN string, letters []mdsend.Letter) (err error) {
	conn, err := sqlite.OpenConn(
		conntectionDSN,
		// sqlite.OpenCreate, sqlite.OpenReadWrite,
	)
	if err != nil {
		return fmt.Errorf("database %q inaccessible: %w", conntectionDSN, err)
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

	for _, letter := range letters {
		if err = queue.CreateLetter(ctx, letter); err != nil {
			return err
		}
		for range 1 {
			err = queue.CreateAttachment(ctx, mdsend.Attachment{
				LetterID:    letter.ID,
				Name:        "random",
				Source:      "",
				Hash:        "",
				ContentID:   "",
				ContentType: "",
				Content:     nil,
			})
			if err != nil {
				return err
			}
		}
		for i := range 100 {
			err = queue.CreateMessage(ctx, mdsend.Message{
				ID: fmt.Sprintf("testMessage%d", i),
				From: mail.Address{
					Name:    "random",
					Address: fmt.Sprintf("testMessage%d@example.com", i),
				},
				To: mail.Address{
					Name:    "random",
					Address: fmt.Sprintf("testAddress%d@example.com", i),
				},
				LetterID:      letter.ID,
				Subject:       fmt.Sprintf("testMessage%d", i),
				Text:          "random",
				HTML:          "",
				ScheduleAfter: time.Time{},
				ScheduledAt:   time.Time{},
				SentAt:        time.Time{},
			})
			if err != nil {
				return err
			}
		}
	}
	return err
}
