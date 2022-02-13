package providers

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"time"

	mailgun "github.com/mailgun/mailgun-go/v4"
)

// NewMailgunProvider creates a Mailgun sending agent.
func NewMailgunProvider(apiURI string) *MailgunProvider {
	r := regexp.MustCompile(`^([^@]+)@(.+)$`)
	m := r.FindStringSubmatch(apiURI)
	if len(m) == 0 {
		log.Fatal(`Mailgun API URI must match format <apikey>@<apidomain>.`)
	}
	return &MailgunProvider{apiKey: m[1], apiDomain: m[2]}
}

// MailgunProvider delivers messages through Mailgun API.
type MailgunProvider struct {
	apiKey, apiDomain string
	mg                *mailgun.MailgunImpl
}

// Open sets up a connection to the Mailgun server.
func (p *MailgunProvider) Open() error {
	p.mg = mailgun.NewMailgun(p.apiDomain, p.apiKey)
	return nil
}

// Close tears down the existing connection.
func (p *MailgunProvider) Close() error {
	return nil
}

// Send queues one message to one recipient.
func (p *MailgunProvider) Send(to string, MIME io.ReadCloser, test bool) (string, error) {
	// log.Fatal(`debugging`)
	message := p.mg.NewMIMEMessage(MIME, to)
	if test {
		log.Println(`Testing mode is on...`)
		message.EnableTestMode()
	}
	// message.EnableNativeSend() // this one fails to deliver mail

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	status, id, err := p.mg.Send(ctx, message)
	return fmt.Sprintf(`%s %s`, status, id), err
}
