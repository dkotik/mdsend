package mailgun

import (
	"github.com/dkotik/mdsend"
	"github.com/mailgun/mailgun-go/v4"
)

func (s mailgunSender) prepareMessage(d mdsend.Dispatch) *mailgun.Message {
	message := s.NewMessage(
		d.From.String(),
		d.Subject,
		d.Text,
		d.To.String(),
	)
	message.SetHtml(d.HTML)
	for _, h := range d.Headers {
		message.AddHeader(h.Name, h.Value)
	}
	return message
}
