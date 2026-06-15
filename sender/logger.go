package sender

import (
	"context"
	"log/slog"

	"github.com/dkotik/mdsend"
)

type logger struct {
	mdsend.Sender
	Logger *slog.Logger
}

func NewLogger(l *slog.Logger) func(mdsend.Sender) mdsend.Sender {
	if l == nil {
		l = slog.Default()
	}
	return func(s mdsend.Sender) mdsend.Sender {
		if s == nil {
			panic("sender is nil")
		}
		return logger{Sender: s, Logger: l}
	}
}

func (l logger) Send(ctx context.Context, msg mdsend.Dispatch) (id string, err error) {
	id, err = l.Sender.Send(ctx, msg)
	if err == nil {
		l.Logger.DebugContext(
			ctx,
			"sent: "+msg.Subject,
			slog.String("id", msg.ID),
			slog.String("letter_id", msg.LetterID),
			slog.String("queue_id", id),
			slog.String("address", msg.From.Address),
		)
	} else {
		l.Logger.ErrorContext(
			ctx,
			msg.Subject,
			slog.String("id", msg.ID),
			slog.String("letter_id", msg.LetterID),
			slog.String("queue_id", id),
			slog.String("address", msg.From.Address),
			slog.String("error", err.Error()),
		)
	}
	return id, err
}
