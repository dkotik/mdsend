package queue

import (
	"context"
	"iter"
	"time"

	"github.com/dkotik/mdsend"
	"golang.org/x/sync/errgroup"
)

type scanner struct {
	Frequency        time.Duration
	LetterLimit      int
	LetterBatch      []mdsend.Letter
	LetterCursor     Cursor
	MessageCursor    ChildCursor
	Queue            Queue
	QueuedMessageIDs chan []string
	Scheduler        Scheduler
}

func NewScanner(
	frequency time.Duration,
	letterCursor Cursor,
	messageCursor ChildCursor,
	scheduler Scheduler,
) (Process, <-chan []string) {
	if scheduler == nil {
		panic("scheduler is nil")
	}
	s := &scanner{
		Frequency:        frequency,
		LetterCursor:     letterCursor,
		MessageCursor:    messageCursor,
		Scheduler:        scheduler,
		QueuedMessageIDs: make(chan []string),
	}
	if s.Frequency == 0 {
		s.Frequency = time.Second
	}
	if s.LetterCursor.Batch == 0 {
		s.LetterCursor.Batch = -5
	}
	if s.MessageCursor.Batch == 0 {
		s.MessageCursor.Batch = -200
	}
	if s.LetterCursor.Batch > 0 {
		s.LetterLimit = int(s.LetterCursor.Batch)
	} else {
		s.LetterLimit = int(-s.LetterCursor.Batch)
	}
	s.LetterBatch = make([]mdsend.Letter, 0, s.LetterLimit)
	return s, s.QueuedMessageIDs
}

func (s *scanner) JoinErrorGroup(ctx context.Context, errGroup *errgroup.Group, q Queue) {
	if q == nil {
		panic("queue is nil")
	}
	if s.Queue != nil {
		panic("scanner is already bound to an error group")
	}
	s.Queue = q
	errGroup.Go(func() error {
		defer close(s.QueuedMessageIDs)
		return s.Scan(ctx)
	})
}

func (s *scanner) MarkLetterAsSent(ctx context.Context, ID string) (err error) {
	q, tx, err := s.Queue.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	_, err = q.MarkLetterAsSent(ctx, ID)
	tx.Close(&err)
	return err
}

func (s *scanner) LoadNextBatchOfUnsentLetters(ctx context.Context) (err error) {
	s.LetterBatch = s.LetterBatch[:0]
	letterPull, letterStop := iter.Pull2[mdsend.Letter, error](s.Queue.ListLetters(ctx, s.LetterCursor))
	for range s.LetterLimit {
		letter, err, ok := letterPull()
		if err != nil {
			letterStop()
			return err
		}
		if !ok {
			break
		}
		if !letter.SentAt.IsZero() {
			// skip letters that have already been sent
			// TODO: try to expire the letter according to schedule
			continue
		}
		s.LetterBatch = append(s.LetterBatch, letter)
	}
	found := len(s.LetterBatch)
	if found == 0 {
		// wrap the cursor to the beginning
		s.LetterCursor.ItemID = ""
	} else {
		// aim cursor to the next batch
		s.LetterCursor.ItemID = s.LetterBatch[found-1].ID
	}
	letterStop()
	return nil
}

func (s *scanner) Scan(ctx context.Context) (err error) {
	foundUnsent := 0
	batchSize := s.MessageCursor.Batch
	if batchSize < 0 {
		batchSize = -batchSize
	}
	batch := make([]string, 0, batchSize)
	messages := make([]mdsend.Message, 0, batchSize)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.Frequency):
		}
		if err = s.LoadNextBatchOfUnsentLetters(ctx); err != nil {
			return err
		}

		for _, letter := range s.LetterBatch {
			s.MessageCursor.ParentID = letter.ID
			batch = batch[:0]
			messages = messages[:0]
			for {
				messagePull, messageStop := iter.Pull2[mdsend.Message, error](s.Queue.ListMessages(ctx, s.MessageCursor))
				for range s.MessageCursor.Batch {
					message, err, ok := messagePull()
					if err != nil {
						messageStop()
						return err
					}
					if !ok {
						s.MessageCursor.Cursor.ItemID = ""
						break
					}
					s.MessageCursor.Cursor.ItemID = message.ID
					if message.SentAt.IsZero() {
						batch = append(batch, message.ID)
						messages = append(messages, message)
					}
				}
				messageStop()

				if len(batch) > 0 {
					if err = s.Scheduler.ScheduleForDelivery(ctx, messages); err != nil {
						return err
					}
					foundUnsent += len(batch)
					select {
					case <-ctx.Done():
						return ctx.Err()
					case s.QueuedMessageIDs <- batch:
						// delivered a batch of discovered messages that were not yet sent
					}
				}

				if s.MessageCursor.Cursor.ItemID == "" {
					break
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(s.Frequency):
				}
			}

			if foundUnsent == 0 {
				if err = s.MarkLetterAsSent(ctx, letter.ID); err != nil {
					return err
				}
			}
			foundUnsent = 0
		}
	}
}
