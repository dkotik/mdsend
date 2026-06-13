package smtp

import (
	"bytes"
	"context"
	"net/smtp"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/mime"
)

func (s senderSMTP) Send(ctx context.Context, m mdsend.Dispatch) (_ string, err error) {
	b := &bytes.Buffer{}
	if err = mime.NewWriter(b, s.Queue, nil).Write(ctx, m); err != nil {
		return "", err
	}

	return m.ID, smtp.SendMail(
		s.Connection,
		s.Authentication,
		m.From.Address,         // must be without a name
		[]string{m.To.Address}, // must be without a name
		b.Bytes(),
	)
}
