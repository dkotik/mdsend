package queue

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"sync"
	"time"

	"github.com/dkotik/mdsend"
	"golang.org/x/exp/maps"
)

var (
	_ Queue          = (*progressTracker)(nil)
	_ Confirmer      = (*progressTracker)(nil)
	_ slog.LogValuer = (*Progress)(nil)
)

type Progress struct {
	Sent    int64
	Total   int64
	Average time.Duration
}

func (p Progress) EstimateRemaining() time.Duration {
	return time.Duration(p.Total-p.Sent) * p.Average
}

func (p Progress) MessagesPerSecond() int64 {
	if p.Average == 0 {
		return 0
	}
	return int64(time.Second / p.Average)
}

func (p Progress) MessagesPerMinute() int64 {
	if p.Average == 0 {
		return 0
	}
	return int64(time.Minute / p.Average)
}

func (p Progress) OfOne() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Sent) / float64(p.Total)
}

func (p Progress) String() string {
	return fmt.Sprintf("%d/%d (%.2f%%)", p.Sent, p.Total, p.OfOne()*100)
}

func (p Progress) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("sent", fmt.Sprintf("%d/%d", p.Sent, p.Total)),
		slog.String("speed", fmt.Sprintf("%dm/s", p.MessagesPerSecond())),
		// slog.String("average", fmt.Sprintf("%.2fs", p.Average.Seconds())),
		slog.String("estimate_remaining", fmt.Sprintf("%.2fs", p.EstimateRemaining().Seconds())),
	)
}

type ProgressTracker interface {
	TrackProgress(context.Context, Progress)
}

type ProgressTrackerFunc func(context.Context, Progress)

func (f ProgressTrackerFunc) TrackProgress(ctx context.Context, p Progress) {
	f(ctx, p)
}

type progressTracker struct {
	Queue
	Report ProgressTracker

	mu               *sync.Mutex
	pendingLetterIDs map[string]struct{}
	pendingMessages  map[string]string
	sentMessages     map[string]struct{}
	lastReportTime   time.Time
	lastReportCount  int64
	fullScanCount    int
}

func NewProgressTracker(
	q Queue,
	report ProgressTracker,
) interface {
	Queue
	Confirmer
} {
	return &progressTracker{
		Queue:            q,
		Report:           report,
		mu:               &sync.Mutex{},
		pendingLetterIDs: make(map[string]struct{}),
		pendingMessages:  make(map[string]string),
		sentMessages:     make(map[string]struct{}),
		lastReportTime:   time.Now(),
	}
}

func (p *progressTracker) announceProgress(ctx context.Context) {
	fresh := int64(len(p.sentMessages))
	if fresh == p.lastReportCount {
		return // there is no new progress to report
	}
	t := time.Now()
	delta := t.Sub(p.lastReportTime)
	report := Progress{
		Sent:    fresh,
		Total:   int64(len(p.pendingMessages)) + fresh,
		Average: delta / time.Duration(fresh-p.lastReportCount),
	}
	p.lastReportTime = t
	p.lastReportCount = fresh
	p.Report.TrackProgress(ctx, report)
}

func (p *progressTracker) MarkMessagesAsScheduled(
	ctx context.Context,
	letterID string,
	IDs ...string,
) (err error) {
	if err = p.Queue.MarkMessagesAsScheduled(ctx, letterID, IDs...); err != nil {
		return err
	}
	p.mu.Lock()
	for _, ID := range IDs {
		p.pendingMessages[ID] = letterID
		// p.pendingMessageCount++
	}
	p.pendingLetterIDs[letterID] = struct{}{}
	p.mu.Unlock()
	return nil
}

func (p *progressTracker) ConfirmScheduling(ctx context.Context, c Confirmation) (err error) {
	p.mu.Lock()
	delete(p.pendingMessages, c.MessageID)
	p.sentMessages[c.MessageID] = struct{}{}
	p.mu.Unlock()
	return nil
}

func (p *progressTracker) MarkMessageAsSent(
	ctx context.Context,
	ID string,
) (ok bool, err error) {
	if ok, err = p.Queue.MarkMessageAsSent(ctx, ID); err != nil {
		return false, err
	}
	p.mu.Lock()
	delete(p.pendingMessages, ID)
	p.sentMessages[ID] = struct{}{}
	p.mu.Unlock()
	return ok, nil
}

func (p *progressTracker) ListLetters(
	ctx context.Context,
	cursor Cursor,
) iter.Seq2[mdsend.Letter, error] {
	return func(yield func(mdsend.Letter, error) bool) {
		var (
			pending             []string
			sent                []string
			letter              mdsend.Letter
			messageID, letterID string
			err                 error
			ID                  string
			ok                  bool
		)
		for letter, err = range p.Queue.ListLetters(ctx, cursor) {
			if letter.SentAt.IsZero() {
				pending = append(pending, letter.ID)
			} else {
				sent = append(sent, letter.ID)
			}
			if !yield(letter, err) {
				break
			}
		}

		p.mu.Lock()
		for _, ID = range sent {
			if _, ok = p.pendingLetterIDs[ID]; ok {
				delete(p.pendingLetterIDs, ID)
				for messageID, letterID = range p.pendingMessages {
					if letterID == ID {
						delete(p.pendingMessages, messageID)
						p.sentMessages[messageID] = struct{}{}
					}
				}
			}
		}
		for _, ID = range pending {
			p.pendingLetterIDs[ID] = struct{}{}
		}
		if cursor.ItemID == "" && len(p.pendingMessages) == 0 {
			p.fullScanCount++
			if p.fullScanCount > 6 {
				maps.Clear(p.sentMessages)
				p.fullScanCount = 0
				p.lastReportCount = 0
			}
		}
		p.announceProgress(ctx)
		p.mu.Unlock()
	}
}

func (p *progressTracker) ListMessages(
	ctx context.Context,
	cursor ChildCursor,
) iter.Seq2[mdsend.Message, error] {
	return func(yield func(mdsend.Message, error) bool) {
		var (
			pending []string
			sent    []string
			// durations      time.Duration
			// durationsCount time.Duration
			// estDuration    = time.Now().Add(p.LastAverage)
			message  mdsend.Message
			letterID string
			ID       string
			ok       bool
			err      error
		)
		for message, err = range p.Queue.ListMessages(ctx, cursor) {
			if !message.ScheduledAt.IsZero() { // skip messages that are not yet scheduled
				if message.SentAt.IsZero() {
					pending = append(pending, message.ID)
					// durations = durations + estDuration.Sub(message.ScheduledAt)
					// durationsCount++
				} else {
					sent = append(sent, message.ID)
					// durations = durations + message.SentAt.Sub(message.ScheduledAt)
					// durationsCount++
				}
			}
			if !yield(message, err) {
				break
			}
		}

		p.mu.Lock()
		for _, ID = range sent {
			if _, ok = p.pendingMessages[ID]; ok {
				delete(p.pendingMessages, ID)
				p.sentMessages[ID] = struct{}{}
			}
		}
		if len(pending) > 0 {
			letterID = cursor.ParentID
			p.pendingLetterIDs[letterID] = struct{}{}
			for _, ID = range pending {
				p.pendingMessages[ID] = letterID
			}
		}
		// if durationsCount > 0 {
		// 	p.LastAverage = durations / durationsCount
		// }
		p.mu.Unlock()
	}
}
