package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/mdsend"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Cursor holds the position of an item in a list from which an iterator
// can retrieve items sequentially. If [Cursor.ItemID] is empty, the iterator
// starts from the first batch. The item with the same ID is
// always skipped, so the iterator will start from the next item.
//
// [Cursor.Batch] sets the maximum number of items to retrieve in one
// repository paging operation. A negative batch value iterates items
// in descending order from the [Cursor.ItemID].
//
// The iterator loads additional batches as needed as long as the range
// of items to retrieve is not exhausted.
//
// Context cancellation will stop the iterator at the end
// of the current batch.
type Cursor struct {
	ItemID string
	Batch  int64
}

type ChildCursor struct {
	ParentID string
	Cursor
}

type Transaction interface {
	Close(*error)
}

type Confirmation struct {
	LetterID       string
	MessageID      string
	ConfirmationID string
	SentAt         time.Time
}

type Queue interface {
	CreateLetter(context.Context, mdsend.Letter) error
	RetrieveLetter(context.Context, string) (mdsend.Letter, error)
	MarkLetterAsSent(context.Context, string) (bool, error)
	DeleteLetter(context.Context, string) error

	CreateAttachment(context.Context, mdsend.Attachment) error
	CreateMessage(context.Context, mdsend.Message) error
	MarkMessagesAsScheduled(context.Context, string, ...string) error
	MarkMessageAsSent(context.Context, string) (bool, error)
	// RetrieveAttachmentContents(context.Context, string) ([]byte, error)

	ListLetters(context.Context, Cursor) iter.Seq2[mdsend.Letter, error]
	ListMessages(context.Context, ChildCursor) iter.Seq2[mdsend.Message, error]
	ListAttachments(context.Context, string) iter.Seq2[mdsend.Attachment, error]

	BeginTransaction(context.Context) (Queue, Transaction, error)
	WithTransaction(context.Context, Transaction) (Queue, error)
}

func NewSender(s mdsend.Mailer) message.HandlerFunc {
	if s == nil {
		panic("sender is nil")
	}
	return func(msg *message.Message) (_ []*message.Message, err error) {
		var m mdsend.Message
		if err = json.Unmarshal(msg.Payload, &m); err != nil {
			return nil, fmt.Errorf("invalid JSON payload: %w", err)
		}
		confirmation := Confirmation{
			LetterID:  m.LetterID,
			MessageID: m.ID,
		}
		// panic("sending")
		confirmation.ConfirmationID, err = s.SendMail(msg.Context(), m)
		if err != nil {
			return nil, err
		}
		confirmation.SentAt = time.Now()
		confirmationBytes, err := json.Marshal(confirmation)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal confirmation: %w", err)
		}
		return []*message.Message{message.NewMessage(uuid.NewString(), confirmationBytes)}, nil
	}
}

// TODO: deprecate?
func MountSenders(
	r *message.Router,
	pub message.Publisher,
	sub message.Subscriber,
	topicPrefix string,
	senders ...mdsend.Mailer,
) {
	if r == nil {
		panic("router is nil")
	}
	if sub == nil {
		panic("subscriber is nil")
	}
	if pub == nil {
		panic("publisher is nil")
	}

	topicPrefix = strings.TrimSpace(topicPrefix)
	if topicPrefix == "" {
		topicPrefix = "mdsend"
	}
	confirmationTopic := fmt.Sprintf("%s_confirmation", topicPrefix)
	for i, s := range senders {
		r.AddHandler(
			fmt.Sprintf("%s_message_sender_%d", topicPrefix, i+1),
			fmt.Sprintf("%s_outbox_%d", topicPrefix, i+1),
			sub,
			confirmationTopic,
			pub,
			// TODO: add retry
			NewSender(s),
		)
	}
}

type Process interface {
	JoinErrorGroup(context.Context, *errgroup.Group, Queue)
}

func CollectMostOf[T any](ctx context.Context, count int) func(iter.Seq2[T, error]) iter.Seq2[T, error] {
	return func(in iter.Seq2[T, error]) iter.Seq2[T, error] {
		return func(yield func(item T, err error) bool) {
			limit := count
			for item, err := range in {
				if !yield(item, err) {
					return
				}
				limit--
				if limit == 0 {
					return
				}
			}
		}
	}
}
