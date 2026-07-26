package mailer

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dkotik/mdsend"
)

func ApplyPrivacyMaskToEmailAddress(email string) string {
	private, domain, ok := strings.Cut(email, "@")
	if !ok {
		return "[INVALID EMAIL ADDRESS]"
	}
	// Mask based on length of local part
	if len(private) <= 2 {
		return "**@" + domain
	}

	// Keep first and last character, mask the middle
	return string(private[0]) + strings.Repeat("*", len(private)-2) + string(private[len(private)-1]) + "@" + domain
}

type logger struct {
	mdsend.Mailer
	Logger       *slog.Logger
	SuccessLevel slog.Level
}

func NewLogger(l *slog.Logger, successLevel slog.Level) func(mdsend.Mailer) mdsend.Mailer {
	if l == nil {
		l = slog.Default()
	}
	return func(s mdsend.Mailer) mdsend.Mailer {
		if s == nil {
			panic("sender is nil")
		}
		return logger{
			Mailer:       s,
			Logger:       l,
			SuccessLevel: successLevel,
		}
	}
}

func (l logger) SendMail(ctx context.Context, msg mdsend.Message) (id string, err error) {
	id, err = l.Mailer.SendMail(ctx, msg)
	address := ApplyPrivacyMaskToEmailAddress(msg.To.Address)
	if err == nil {
		l.Logger.Log(
			ctx,
			l.SuccessLevel,
			"sent: "+msg.Subject,
			slog.String("id", msg.ID),
			slog.String("letter_id", msg.LetterID),
			slog.String("confirmation_id", id),
			slog.String("address", address),
		)
	} else {
		l.Logger.ErrorContext(
			ctx,
			msg.Subject,
			slog.String("id", msg.ID),
			slog.String("letter_id", msg.LetterID),
			slog.String("confirmation_id", id),
			slog.String("address", address),
			slog.String("error", err.Error()),
		)
	}
	return id, err
}
