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

func NewSender(s mdsend.Sender) message.HandlerFunc {
	if s == nil {
		panic("sender is nil")
	}
	return func(msg *message.Message) (_ []*message.Message, err error) {
		var m mdsend.Dispatch
		if err = json.Unmarshal(msg.Payload, &m); err != nil {
			return nil, fmt.Errorf("invalid JSON payload: %w", err)
		}
		confirmation := Confirmation{
			LetterID:  m.LetterID,
			MessageID: m.ID,
		}
		confirmation.ConfirmationID, err = s.Send(msg.Context(), m)
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

func MountSenders(
	r *message.Router,
	pub message.Publisher,
	sub message.Subscriber,
	topicPrefix string,
	senders ...mdsend.Sender,
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
	JoinErrorGroup(context.Context, *errgroup.Group, mdsend.Queue)
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
