package mailer

import (
	"context"
	"time"

	"github.com/dkotik/mdsend"
)

type timeout struct {
	Mailer mdsend.Mailer
	Limit  time.Duration
}

func NewTimeout(d time.Duration) func(mdsend.Mailer) mdsend.Mailer {
	if d < time.Millisecond*10 {
		panic("timeout is less than 10ms")
	}
	return func(m mdsend.Mailer) mdsend.Mailer {
		if m == nil {
			panic("sender is nil")
		}
		return timeout{
			Mailer: m,
			Limit:  d,
		}
	}
}

func (t timeout) SendMail(ctx context.Context, m mdsend.Message) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, t.Limit)
	defer cancel()
	return t.Mailer.SendMail(ctx, m)
}
