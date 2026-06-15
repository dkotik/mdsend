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
	LetterCursor     mdsend.Cursor
	MessageCursor    mdsend.ChildCursor
	Queue            mdsend.Queue
	QueuedMessageIDs chan []string
}

func NewScanner(
	frequency time.Duration,
	letterCursor mdsend.Cursor,
	messageCursor mdsend.ChildCursor,
) (Process, <-chan []string) {
	s := scanner{
		Frequency:        frequency,
		LetterCursor:     letterCursor,
		MessageCursor:    messageCursor,
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
	return s, s.QueuedMessageIDs
}

func (s scanner) JoinErrorGroup(ctx context.Context, errGroup *errgroup.Group, q mdsend.Queue) {
	if q == nil {
		panic("queue is nil")
	}
	if s.Queue != nil {
		panic("scanner is already bound to an error group")
	}
	s.Queue = q
	errGroup.Go(func() error {
		defer close(s.QueuedMessageIDs)
		return s.scan(ctx)
	})
}

func (s scanner) scan(ctx context.Context) (err error) {
	foundUnsent := 0
	for {
		for {
			letterPull, letterStop := iter.Pull2[mdsend.Letter, error](s.Queue.ListLetters(ctx, s.LetterCursor))
			letter, err, ok := letterPull()
			letterStop()
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			if !letter.SentAt.IsZero() {
				continue // skip letters that have already been sent
			}

			foundUnsent = 0
			s.LetterCursor.ItemID = letter.ID
			s.MessageCursor.ParentID = letter.ID

			for {
				batch := make([]string, 0, s.MessageCursor.Batch)
				messagePull, messageStop := iter.Pull2[mdsend.Dispatch, error](s.Queue.ListDispatches(ctx, s.MessageCursor))
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
					}
				}
				messageStop()

				if len(batch) > 0 {
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
		}

		if foundUnsent == 0 {
			// TODO: attempt to mark the letter as sent
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.Frequency):
			s.LetterCursor.ItemID = "" // start the scan over
		}
	}
}
