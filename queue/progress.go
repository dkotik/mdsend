package queue

import (
	"context"
	"fmt"
	"iter"
	"sync"

	"github.com/dkotik/mdsend"
)

var _ Queue = (*progressTracker)(nil)

type Progress struct {
	Sent  int64
	Total int64
}

type ProgressTracker interface {
	TrackProgress(context.Context, Progress)
}

func (p Progress) OfOne() float64 {
	return float64(p.Sent) / float64(p.Total)
}

func (p Progress) String() string {
	return fmt.Sprintf("%d/%d (%.2f%%)", p.Sent, p.Total, p.OfOne()*100)
}

type progressTracker struct {
	Queue
	Report ProgressTracker

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
		mu:               &sync.Mutex{},
		pendingLetterIDs: make(map[string]struct{}),
		pendingMessages:  make(map[string]string),
	}
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
				return
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
			pending  []string
			sent     []string
			message  mdsend.Message
			letterID string
			ID       string
			ok       bool
			err      error
		)
		for message, err = range p.Queue.ListMessages(ctx, cursor) {
			if message.SentAt.IsZero() {
				pending = append(pending, message.ID)
			} else {
				sent = append(sent, message.ID)
			}
			if !yield(message, err) {
				return
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
		p.Report.TrackProgress(ctx, Progress{
			Sent:  p.sentMessageCount,
			Total: p.pendingMessageCount,
		})
		p.mu.Unlock()
	}
}
