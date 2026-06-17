package mailer

import (
	"context"
	"time"

	"github.com/dkotik/mdsend"
)

type delay struct {
	mdsend.Mailer
	Duration time.Duration
}

func NewDelay(d time.Duration) func(mdsend.Mailer) mdsend.Mailer {
	if d < time.Millisecond {
		panic("duration is too short")
	}
	return func(s mdsend.Mailer) mdsend.Mailer {
		if s == nil {
			panic("sender is nil")
		}
		return delay{
			Mailer:   s,
			Duration: d,
		}
	}
}
func (d delay) SendMail(ctx context.Context, msg mdsend.Message) (string, error) {
	select {
	case <-time.After(d.Duration):
		return d.Mailer.SendMail(ctx, msg)
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
