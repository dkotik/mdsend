package sender

import (
	"context"
	"time"

	"github.com/dkotik/mdsend"
)

type delay struct {
	mdsend.Sender
	Duration time.Duration
}

func NewDelay(d time.Duration) func(mdsend.Sender) mdsend.Sender {
	if d < time.Millisecond {
		panic("duration is too short")
	}
	return func(s mdsend.Sender) mdsend.Sender {
		if s == nil {
			panic("sender is nil")
		}
		return delay{
			Sender:   s,
			Duration: d,
		}
	}
}
func (d delay) Send(ctx context.Context, msg mdsend.Dispatch) (string, error) {
	select {
	case <-time.After(d.Duration):
		return d.Sender.Send(ctx, msg)
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
