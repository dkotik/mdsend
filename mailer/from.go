package mailer

import (
	"context"
	"net/mail"

	"github.com/dkotik/mdsend"
)

type fromOverride struct {
	mdsend.Mailer
	mail.Address
}

func NewFromOverride(a mail.Address) func(mdsend.Mailer) mdsend.Mailer {
	return func(m mdsend.Mailer) mdsend.Mailer {
		if m == nil {
			panic("mailer is nil")
		}
		return fromOverride{
			Mailer:  m,
			Address: a,
		}
	}
}

func (f fromOverride) SendMail(ctx context.Context, m mdsend.Message) (string, error) {
	m.From = f.Address
	return f.Mailer.SendMail(ctx, m)
}
