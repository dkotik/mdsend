package queue

import (
	"context"
	"encoding/json"
	"fmt"
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

func (p Progress) OfOne() float64 {
	return float64(p.Sent) / float64(p.Total)
}

func (p Progress) String() string {
	return fmt.Sprintf("%d/%d (%.2f%%)", p.Sent, p.Total, p.OfOne()*100)
}

type progressTracker struct {
	Queue      mdsend.Queue
	Discovered <-chan []string
	Sent       chan string
	Progress   chan Progress
}

func NewProgressTracker(
	pendingMessageIDs <-chan []string,
) (Process, message.NoPublishHandlerFunc, <-chan Progress) {
	tracker := &progressTracker{
		Discovered: pendingMessageIDs,
		Sent:       make(chan string),
		Progress:   make(chan Progress), // closed by Run
	}
	return tracker, tracker.Handle, tracker.Progress
}

func (t *progressTracker) JoinErrorGroup(ctx context.Context, eg *errgroup.Group, q mdsend.Queue) {
	if q == nil {
		panic("queue is nil")
	}
	if t.Queue != nil {
		panic("tracker is already bound to an error group")
	}
	t.Queue = q
	eg.Go(func() error {
		defer close(t.Progress)
		return t.Run(ctx)
	})
}

func (t *progressTracker) Run(ctx context.Context) (err error) {
	progress := Progress{}
	known := make(map[string]bool)
	update, ok := false, false
	id := ""
	for {
		select {
		case <-ctx.Done():
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

func (t *progressTracker) Handle(msg *message.Message) (err error) {
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
