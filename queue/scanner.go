package queue

import (
	"context"
	"iter"
	"time"

	"github.com/dkotik/mdsend"
	"golang.org/x/sync/errgroup"
)

type ContinuousScannerOptions struct {
	Frequency             time.Duration
	ProgressTracker       ProgressTracker
	LetterBatchSize       uint8
	MessageBatchSize      uint16
	BeginWithOlderLetters bool
}

type continuousScanner struct {
	Queue     Queue
	Scheduler Scheduler
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
	if options.ProgressTracker != nil {
		q = NewProgressTracker(q, options.ProgressTracker)
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
	cs := continuousScanner{
		Queue:     q,
		Scheduler: s,
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
				// panic(fmt.Sprintf("%s: %+v", letter.ID, lc))
				letterStop()
				return err
			}
			if !ok {
				// wrap the cursor to the beginning
				lc.ItemID = ""
				break
			}
			if !letter.SentAt.IsZero() {
				// skip letters that have already been sent
				// TODO: try to expire the letter according to schedule
				continue
			}
			letters = append(letters, letter)
		}
		letterStop()
		if ok {
			// aim cursor to the next batch
			lc.ItemID = letter.ID
		}

		for _, letter := range letters {
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
					if message.SentAt.IsZero() {
						messages = append(messages, message)
					}
				}
				messageStop()

				if foundUnsentInBatch = len(messages); foundUnsentInBatch > 0 {
					if err = s.Scheduler.ScheduleForDelivery(ctx, messages); err != nil {
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
