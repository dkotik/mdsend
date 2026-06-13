package mailgun

import (
	"bytes"
	"context"
	"io"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/mime"
	"github.com/mailgun/mailgun-go/v4"
)

func (s mailgunSender) prepareMessage(
	ctx context.Context,
	d mdsend.Dispatch,
) (_ *mailgun.Message, err error) {
	b := &bytes.Buffer{}
	if err = mime.NewWriter(b, s.Queue, nil).Write(ctx, d); err != nil {
		return nil, err
	}
	message := s.NewMIMEMessage(io.NopCloser(b), d.To.String())

	// message := s.NewMessage(
	// 	d.From.String(),
	// 	d.Subject,
	// 	d.Text,
	// 	d.To.String(),
	// )
	// message.SetHtml(d.HTML)
	// for _, h := range d.Headers {
	// 	message.AddHeader(h.Name, h.Value)
	// }
	return message, nil
}
