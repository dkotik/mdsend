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
	Discovered chan []string
	Sent       chan string
	Progress   chan Progress
}

func NewProgressTracker(
	ctx context.Context,
	queue mdsend.Queue,
	pendingMessageIDs chan []string,
	dependencies *errgroup.Group,
) (message.NoPublishHandlerFunc, chan Progress) {
	tracker := progressTracker{
		Queue:      queue,
		Discovered: pendingMessageIDs,
		Sent:       make(chan string),
		Progress:   make(chan Progress), // closed by Run
	}
	dependencies.Go(func() error {
		return tracker.Run(ctx)
	})
	return tracker.Handle, tracker.Progress
}

func (t progressTracker) Run(ctx context.Context) (err error) {
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
		// case <-ticker.C:
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
