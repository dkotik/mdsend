package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/address"
	"github.com/dkotik/mdsend/internal/media"
	"github.com/dkotik/mdsend/internal/template"
	"github.com/dkotik/mdsend/queue"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"github.com/oklog/ulid/v2"
	"github.com/urfave/cli/v3"
)

func cmdQueueAdd(ctx context.Context, c *cli.Command) (err error) {
	if !c.Args().Present() {
		return errors.New(`no Markdown letters selected to add`)
	}
	fs := media.NewUnsafeUnconstrainedFileSystem()
	p := c.Args().First()
	letter, err := mdsend.NewLetterFromFile(ctx, fs, p)
	if err != nil {
		return fmt.Errorf(`unable to parse letter from file %q: %w`, p, err)
	}

	var (
		connectionDSN            string
		connectionFileDescriptor os.FileInfo
	)
	if c.IsSet(flagQueue.Name) {
		connectionDSN = c.String(flagQueue.Name)
	} else {
		connectionDSN = letter.GetDatabase()
		if connectionDSN == "" {
			connectionDSN = c.String(flagQueue.Name)
		} else {
			connectionFileDescriptor, err = os.Stat(connectionDSN)
			if err != nil {
				return fmt.Errorf("database file %q inaccessible: %w", connectionDSN, err)
			}
		}
	}

	conn, err := newQueueConnection(connectionDSN)
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

	logger := getLogger(c)
	if _, err = queueLetter(
		ctx,
		queue,
		letter,
		p,
		fs,
		logger,
	); err != nil {
		return fmt.Errorf(
			"unable to queue letter: %s: %w",
			filepath.Base(p),
			err,
		)
	}

	for _, p = range c.Args().Slice()[1:] {
		letter, err := mdsend.NewLetterFromFile(ctx, fs, p)
		if err != nil {
			return fmt.Errorf(`unable to parse letter from file %q: %w`, p, err)
		}
		if connectionFileDescriptor != nil {
			if database := letter.GetDatabase(); database != "" {
				alterantiveFileDescriptor, err := os.Stat(database)
				if err != nil {
					return fmt.Errorf("database file %q inaccessible: %w", database, err)
				}
				if !os.SameFile(
					connectionFileDescriptor,
					alterantiveFileDescriptor,
				) {
					return fmt.Errorf("cannot add letters that use different databases in one atomic operation: %q vs %q; override database location with --database flag", connectionDSN, database)
				}
			}
		}
		if _, err = queueLetter(
			ctx,
			queue,
			letter,
			p,
			fs,
			logger,
		); err != nil {
			return fmt.Errorf(
				"unable to queue letter: %s: %w",
				filepath.Base(p),
				err,
			)
		}
	}
	return nil
}

func queueLetter(
	ctx context.Context,
	q queue.Queue,
	letter mdsend.Letter,
	letterPath string,
	fs fs.FS,
	logger *slog.Logger,
) (queued int, err error) {
	if letter.ID == "" {
		letter.ID = ulid.Make().String()
	}
	tmpl, err := template.New(letter, template.Options{})
	if err != nil {
		return queued, err
	}
	if err = q.CreateLetter(ctx, letter); err != nil {
		return queued, err
	}
	rootDirectory := filepath.Dir(letterPath)
	constraints, err := letter.GetMediaConstraints()
	if err != nil {
		return queued, err
	}
	for src := range letter.EachAttachmentSource() {
		attachment, err := mdsend.NewAttachmentFromFile(
			fs,
			src.Location,
			constraints,
		)
		if err != nil {
			return queued, err
		}
		attachment.LetterID = letter.ID
		attachment.Name = src.Name
		if err = q.CreateAttachment(ctx, attachment); err != nil {
			return queued, err
		}
	}

	for recipient, err := range address.Each(
		letter.Frontmatter,
		rootDirectory,
		fs,
	) {
		if err != nil {
			return queued, err
		}

		email, _ := recipient[address.FieldEmail].(string)
		if email == "" {
			return queued, address.ErrAbsentEmailAddress
		}

		message, err := tmpl.RenderLetterForRecipient(recipient)
		if err != nil {
			if queue.IsSkipSentinelError(err) {
				// template indicated that message should be skipped
				continue
			}
			return queued, err
		}
		if err = q.CreateMessage(ctx, message); err != nil {
			if errors.Is(err, mdsend.ErrDuplicateMessage) {
				logger.Warn(
					"message has already been sent",
					slog.String("letter_id", letter.ID),
					slog.String("message_id", message.ID),
					slog.String("seed_key", message.SeedKey),
				)
				continue // ignore duplicate messages
			}
			return queued, err
		}
		queued++
	}

	if queued == 0 {
		return queued, errors.New("letter has no recipients")
	}
	return queued, nil
}
