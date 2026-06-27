package queue

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"sync"
	"time"

	"github.com/dkotik/mdsend"
)

var (
	_ Queue          = (*progressTracker)(nil)
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
		slog.Int64("sent", p.Sent),
		slog.Int64("pending", p.Total),
		slog.String("speed", fmt.Sprintf("%dm/min", p.MessagesPerMinute())),
		slog.String("average", fmt.Sprintf("%.2fs", p.Average.Seconds())),
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
	Report      ProgressTracker
	LastAverage time.Duration

	mu                  *sync.Mutex
	pendingLetterIDs    map[string]struct{}
	pendingMessages     map[string]string
	pendingMessageCount int64
	sentMessageCount    int64
}

func NewProgressTracker(
	q Queue,
	report ProgressTracker,
) Queue {
	return &progressTracker{
		Queue:            q,
		Report:           report,
		LastAverage:      time.Second,
		mu:               &sync.Mutex{},
		pendingLetterIDs: make(map[string]struct{}),
		pendingMessages:  make(map[string]string),
	}
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
		p.pendingMessageCount++
	}
	p.pendingLetterIDs[letterID] = struct{}{}
	p.mu.Unlock()
	return nil
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
						p.sentMessageCount++
					}
				}
			}
		}
		for _, ID = range pending {
			p.pendingLetterIDs[ID] = struct{}{}
		}
		p.mu.Unlock()
	}
}

func (p progressTracker) ListMessages(
	ctx context.Context,
	cursor ChildCursor,
) iter.Seq2[mdsend.Message, error] {
	return func(yield func(mdsend.Message, error) bool) {
		var (
			pending        []string
			sent           []string
			durations      time.Duration
			durationsCount time.Duration
			estDuration    = time.Now().Add(p.LastAverage)
			message        mdsend.Message
			letterID       string
			ID             string
			ok             bool
			err            error
		)
		for message, err = range p.Queue.ListMessages(ctx, cursor) {
			if message.ScheduledAt.IsZero() {
				continue // skip messages that are not yet scheduled
			}
			if message.SentAt.IsZero() {
				pending = append(pending, message.ID)
				durations = durations + estDuration.Sub(message.ScheduledAt)
				durationsCount++
			} else {
				sent = append(sent, message.ID)
				durations = durations + message.SentAt.Sub(message.ScheduledAt)
				durationsCount++
			}
			if !yield(message, err) {
				break
			}
		}

		p.mu.Lock()
		for _, ID = range sent {
			if _, ok = p.pendingMessages[ID]; ok {
				delete(p.pendingMessages, ID)
				p.sentMessageCount++
			}
		}
		if len(pending) > 0 {
			letterID = cursor.ParentID
			p.pendingLetterIDs[letterID] = struct{}{}
			for _, ID = range pending {
				if _, ok = p.pendingMessages[ID]; !ok {
					p.pendingMessages[ID] = letterID
					p.pendingMessageCount++
				}
			}
		}
		if durationsCount > 0 {
			p.LastAverage = durations / durationsCount
		}
		p.Report.TrackProgress(ctx, Progress{
			Sent:    p.sentMessageCount,
			Total:   p.pendingMessageCount,
			Average: p.LastAverage,
		})
		p.mu.Unlock()
	}
}
