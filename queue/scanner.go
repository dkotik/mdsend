package queue

import (
	"context"
	"errors"
	"iter"
	"log/slog"
	"time"

	"github.com/dkotik/mdsend"
	"golang.org/x/sync/errgroup"
)

type ContinuousScannerOptions struct {
	Frequency             time.Duration
	LetterBatchSize       uint8
	MessageBatchSize      uint16
	BeginWithOlderLetters bool
	Logger                *slog.Logger
}

type continuousScanner struct {
	Queue     Queue
	Scheduler Scheduler
	Logger    *slog.Logger
}

func NewContinuousScanner(
	ctx context.Context,
	wg *errgroup.Group,
	q Queue,
	s Scheduler,
	options ContinuousScannerOptions,
) {
	if ctx == nil {
		panic("context is nil")
	}
	if wg == nil {
		panic("error group is nil")
	}
	if q == nil {
		panic("queue is nil")
	}
	if s == nil {
		panic("scheduler is nil")
	}
	if options.Frequency < time.Millisecond*20 {
		options.Frequency = time.Second
	}
	if options.LetterBatchSize == 0 {
		options.LetterBatchSize = 10
	}
	if options.MessageBatchSize == 0 {
		options.MessageBatchSize = 200
	}
	letterCursor := Cursor{
		ItemID: "",
		Batch:  -int64(options.LetterBatchSize),
	}
	messageCursor := ChildCursor{
		ParentID: "",
		Cursor: Cursor{
			ItemID: "",
			Batch:  -int64(options.MessageBatchSize),
		},
	}
	if options.BeginWithOlderLetters {
		letterCursor.Batch *= -1
		messageCursor.Batch *= -1
	}
	if options.Logger == nil {
		options.Logger = slog.Default()
	}
	cs := continuousScanner{
		Queue:     q,
		Scheduler: s,
		Logger:    options.Logger,
	}
	pulse := time.NewTicker(options.Frequency).C
	wg.Go(func() error {
		return cs.Scan(ctx, letterCursor, messageCursor, pulse)
	})
}

func (s continuousScanner) MarkLetterAsSent(ctx context.Context, ID string) (err error) {
	q, tx, err := s.Queue.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	_, err = q.MarkLetterAsSent(ctx, ID)
	tx.Close(&err)
	return err
}

func (s continuousScanner) Scan(
	ctx context.Context,
	lc Cursor,
	mc ChildCursor,
	pulse <-chan time.Time,
) (err error) {
	letterBatchSize := lc.Batch
	if letterBatchSize < 0 {
		letterBatchSize = -letterBatchSize
	}
	messageBatchSize := mc.Batch
	if messageBatchSize < 0 {
		messageBatchSize = -messageBatchSize
	}
	letters := make([]mdsend.Letter, 0, letterBatchSize)
	messages := make([]mdsend.Message, 0, messageBatchSize)
	foundUnsentInLetter := 0
	foundUnsentInBatch := 0
	letter := mdsend.Letter{}
	message := mdsend.Message{}
	ok := false

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pulse:
		}

		letters = letters[:0]
		letterPull, letterStop := iter.Pull2[mdsend.Letter, error](s.Queue.ListLetters(ctx, lc))
		for range letterBatchSize {
			letter, err, ok = letterPull()
			if err != nil {
				letterStop()
				return err
			}
			if !ok {
				// wrap the cursor to the beginning
				lc.ItemID = ""
				break
			}
			letters = append(letters, letter)
		}
		letterStop()
		if ok {
			// aim cursor to the next batch
			lc.ItemID = letter.ID
		}

		for _, letter := range letters {
			if !letter.SentAt.IsZero() {
				expiration, err := letter.GetExpiration()
				if err == nil {
					if letter.SentAt.Add(expiration).Before(time.Now()) {
						subject, _ := letter.GetSubject()
						if err = s.Queue.DeleteLetter(ctx, letter.ID); err != nil {
							s.Logger.Error(
								"unable to expire and delete letter",
								slog.String("letter_id", letter.ID),
								slog.String("subject", subject),
								slog.Any("error", err),
							)
						} else {
							s.Logger.Info(
								"removed expired letter from the database",
								slog.String("letter_id", letter.ID),
								slog.String("subject", subject),
							)
						}
					}
				} else if !errors.Is(err, mdsend.ErrFieldNotFound) {
					subject, _ := letter.GetSubject()
					s.Logger.Error(
						"unable to obtain letter expiration",
						slog.String("letter_id", letter.ID),
						slog.String("subject", subject),
						slog.Any("error", err),
					)
				}
				// skip letters that have already been sent
				continue
			}

			mc.ParentID = letter.ID
			messages = messages[:0]
			mc.Cursor.ItemID = ""
			foundUnsentInLetter = 0

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-pulse:
				}

				messagePull, messageStop := iter.Pull2[mdsend.Message, error](s.Queue.ListMessages(ctx, mc))
				for range messageBatchSize {
					message, err, ok = messagePull()
					if err != nil {
						messageStop()
						return err
					}
					if !ok {
						break
					}
					if message.ScheduledAt.IsZero() {
						messages = append(messages, message)
					}
				}
				messageStop()
				// fmt.Println("---------- found messages:", messages, mc)

				if foundUnsentInBatch = len(messages); foundUnsentInBatch > 0 {
					if err = s.Scheduler.ScheduleForDelivery(ctx, letter, messages); err != nil {
						return err
					}
					foundUnsentInLetter += foundUnsentInBatch
				}
				if !ok {
					break
				}
				mc.Cursor.ItemID = message.ID
			}

			if foundUnsentInLetter == 0 {
				if err = s.MarkLetterAsSent(ctx, letter.ID); err != nil {
					return err
				}
			}
		}
	}
}
