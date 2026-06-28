package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Confirmation struct {
	ID        string
	LetterID  string
	MessageID string
	SentAt    time.Time
}

type Confirmer interface {
	ConfirmScheduling(context.Context, Confirmation) error
}

type ConfirmerFunc func(context.Context, Confirmation) error

func (f ConfirmerFunc) ConfirmScheduling(ctx context.Context, c Confirmation) error {
	return f(ctx, c)
}

func NewConfirmationHandler(
	q Queue,
	s Confirmer,
	m Marshaler,
) message.NoPublishHandlerFunc {
	if q == nil {
		panic("queue is nil")
	}
	if s == nil {
		panic("confirmer is nil")
	}
	if m == nil {
		panic("marshaler is nil")
	}
	return message.NoPublishHandlerFunc(func(msg *message.Message) (err error) {
		var confirmation Confirmation
		if err = m.UnmarshalMessage(msg, &confirmation); err != nil {
			return fmt.Errorf("invalid confirmation message: %w", err)
		}
		ctx := msg.Context()
		if err = s.ConfirmScheduling(ctx, confirmation); err != nil {
			return fmt.Errorf("confirmation blocked: %w", err)
		}
		_, err = q.MarkMessageAsSent(ctx, confirmation.MessageID)
		return err
	})
}
