package mailer

import (
	"context"
	"log/slog"

	"github.com/dkotik/mdsend"
)

type logger struct {
	mdsend.Mailer
	Logger *slog.Logger
}

func NewLogger(l *slog.Logger) func(mdsend.Mailer) mdsend.Mailer {
	if l == nil {
		l = slog.Default()
	}
	return func(s mdsend.Mailer) mdsend.Mailer {
		if s == nil {
			panic("sender is nil")
		}
		return logger{Mailer: s, Logger: l}
	}
}

func (l logger) SendMail(ctx context.Context, msg mdsend.Message) (id string, err error) {
	id, err = l.Mailer.SendMail(ctx, msg)
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
