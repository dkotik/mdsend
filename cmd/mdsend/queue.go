package main

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/media"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"github.com/urfave/cli/v3"
)

func cmdQueueAdd(ctx context.Context, c *cli.Command) (err error) {
	if c.Args().Len() == 0 {
		if err = addLetters(ctx, c.String(flagDatabase.Name), []mdsend.Letter{
			mdsend.Letter{
				ID: "firstTestLetter" + fmt.Sprintf("%d", time.Now().UnixNano()),
			},
		}); err != nil {
			return fmt.Errorf(`unable to add test letter: %w`, err)
		}
		return errors.New(`no letters selected to add`)
	}
	// panic("woo")
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
	// if !c.IsSet(flagDatabase.Name) {
	// 	// TODO: this is not needed as long as transactions are applied properly
	// 	alternativeQueueFile := letters[0].GetDatabase()
	// 	if alternativeQueueFile != "" {
	// 		p = alternativeQueueFile
	// 	}
	// 	for _, letter := range letters[1:] {
	// 		if letter.GetDatabase() != "" && letter.GetDatabase() != p {
	// 			return fmt.Errorf(`atomic operations require all letters to have the same queue: %q vs %q`, letter.GetDatabase(), p)
	// 		}
	// 	}
	// }
	return addLetters(ctx, p, letters)
}

func addLetters(ctx context.Context, connectionDSN string, letters []mdsend.Letter) (err error) {
	conn, err := newDatabaseConnection(connectionDSN)
	if err != nil {
		return fmt.Errorf("database %q inaccessible: %w", connectionDSN, err)
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
			if errors.Is(err, mdsend.ErrDuplicateLetter) {
				err = nil
				continue // already populated
			}
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
				ID: fmt.Sprintf("testMessage%d%s", i, letter.ID),
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
