package mailer

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/dkotik/mdsend"
)

var ErrLimitExceeded = errors.New("limit exceeded")

type limitedMailer struct {
	mdsend.Mailer
	Remaining   *atomic.Int32
	GracePeriod time.Duration
}

func NewLimitMailer(m mdsend.Mailer, limit uint32, gracePeriod time.Duration) limitedMailer {
	l := &atomic.Int32{}
	l.Add(int32(limit))
	return limitedMailer{
		Mailer:      m,
		Remaining:   l,
		GracePeriod: gracePeriod,
	}
}

func (lm limitedMailer) SendMail(ctx context.Context, m mdsend.Message) (string, error) {
	v := lm.Remaining.Add(-1)
	if v < 0 {
		// grace period before returning error
		<-time.After(lm.GracePeriod)
		return "", ErrLimitExceeded
	}

	return lm.Mailer.SendMail(ctx, m)
}
