package queue

import (
	"context"
	"encoding/json"
	"iter"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/mdsend"
	"golang.org/x/sync/errgroup"
)

type Confirmation struct {
	LetterID       string
	MessageID      string
	ConfirmationID string
	SentAt         time.Time
}

type Progress struct {
	Sent  int
	Total int
}

type progressTracker struct {
	Queue      mdsend.Queue
	Discovered chan []string
	Sent       chan string
	Progress   chan Progress
}

func NewProgressTracker(
	ctx context.Context,
	queue mdsend.Queue,
	frequency time.Duration,
	batchSize int64,
	dependencies *errgroup.Group,
) (message.NoPublishHandlerFunc, chan Progress) {
	if frequency == 0 {
		panic("frequency must be greater than 0")
	}
	if batchSize < 1 {
		panic("batch size must be at least 1")
	}
	tracker := progressTracker{
		Queue:      queue,
		Discovered: make(chan []string),
		Sent:       make(chan string),
		Progress:   make(chan Progress), // closed by Run
	}
	// eg, ctx := errgroup.WithContext(ctx)
	dependencies.Go(func() error {
		return tracker.Run(ctx)
	})
	dependencies.Go(func() error {
		return tracker.Scan(ctx, frequency*3/4, batchSize)
	})
	return tracker.Handle, tracker.Progress
}

func (t progressTracker) Run(ctx context.Context) (err error) {
	// ticker := time.NewTicker(frequency)
	progress := Progress{}
	known := make(map[string]bool)
	update, ok := false, false
	id := ""
	for {
		select {
		case <-ctx.Done():
			close(t.Progress)
			return ctx.Err()
		case batch := <-t.Discovered:
			for _, id = range batch {
				if _, ok = known[id]; !ok {
					known[id] = false
					update = true
				}
			}
		case id = <-t.Sent:
			known[id] = true
			update = true
		case t.Progress <- progress:
			// case <-ticker.C:
			if update {
				update = false
				progress.Total = len(known)
				progress.Sent = 0
				for _, ok = range known {
					if ok {
						progress.Sent++
					}
				}
			}
		}
	}
}

func (t progressTracker) Scan(ctx context.Context, frequency time.Duration, batchSize int64) (err error) {
	nextLetter := mdsend.Cursor{
		ItemID: "",
		Batch:  1,
	}
	nextMessage := mdsend.ChildCursor{
		ParentID: "",
		Cursor: mdsend.Cursor{
			ItemID: "",
			Batch:  batchSize,
		},
	}

	for {
		for {
			letterPull, letterStop := iter.Pull2[mdsend.Letter, error](t.Queue.ListLetters(ctx, nextLetter))
			letter, err, ok := letterPull()
			letterStop()
			if err != nil {
				return err
			}
			if !ok {
				break
			}

			nextLetter.ItemID = letter.ID
			nextMessage.ParentID = letter.ID

			for {
				batch := make([]string, 0, nextMessage.Batch)
				messagePull, messageStop := iter.Pull2[mdsend.Dispatch, error](t.Queue.ListDispatches(ctx, nextMessage))
				for range nextMessage.Batch {
					message, err, ok := messagePull()
					if err != nil {
						messageStop()
						return err
					}
					if !ok {
						nextMessage.Cursor.ItemID = ""
						break
					}
					nextMessage.Cursor.ItemID = message.ID
					if message.SentAt.IsZero() {
						batch = append(batch, message.ID)
					}
				}
				messageStop()
				select {
				case <-ctx.Done():
					return ctx.Err()
				case t.Discovered <- batch:
					// delivered a batch of discovered messages that were not yet sent
				}

				if nextMessage.Cursor.ItemID == "" {
					break
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(frequency):
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(frequency):
			nextLetter.ItemID = "" // start the scan over
		}
	}
}

func (t progressTracker) Handle(msg *message.Message) (err error) {
	var confirmation Confirmation
	if err = json.Unmarshal(msg.Payload, &confirmation); err == nil {
		ctx := msg.Context()
		if err = t.Queue.CompleteDispatch(ctx, confirmation.MessageID); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case t.Sent <- confirmation.MessageID:
		}
	}
	return nil
}
