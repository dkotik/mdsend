package mailer

import (
	"context"
	"math/rand"
	"time"

	"github.com/dkotik/mdsend"
)

type delay struct {
	mdsend.Mailer
	Duration  time.Duration
	Fluctuate time.Duration
}

func NewDelay(t, fluctuate time.Duration) func(mdsend.Mailer) mdsend.Mailer {
	if t < time.Millisecond {
		panic("duration is too short")
	}
	return func(s mdsend.Mailer) mdsend.Mailer {
		if s == nil {
			panic("sender is nil")
		}
		if t+fluctuate < time.Millisecond {
			return s
		}
		return delay{
			Mailer:    s,
			Duration:  t,
			Fluctuate: fluctuate,
		}
	}
}
func (d delay) SendMail(ctx context.Context, msg mdsend.Message) (string, error) {
	delay := d.Duration
	if d.Fluctuate > 0 {
		delay += time.Duration(rand.NormFloat64() * float64(d.Fluctuate))
	}
	select {
	case <-time.After(delay):
		return d.Mailer.SendMail(ctx, msg)
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
