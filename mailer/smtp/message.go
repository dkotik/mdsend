package smtp

import (
	"context"
	"net/smtp"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/mime"
)

func (s senderSMTP) SendMail(ctx context.Context, m mdsend.Message) (_ string, err error) {
	defer s.Buffer.Reset()
	if err = mime.NewWriter(s.Queue, nil).Write(ctx, s.Buffer, m); err != nil {
		return "", err
	}

	return m.ID, smtp.SendMail(
		s.Connection,
		s.Authentication,
		m.From.Address,         // must be without a name
		[]string{m.To.Address}, // must be without a name
		s.Buffer.Bytes(),
	)
}
