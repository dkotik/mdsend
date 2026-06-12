package mailgun

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/dkotik/mdsend"
	mailgun "github.com/mailgun/mailgun-go/v4"
)

const (
	EnvironmentKey    = "MG_API_KEY"
	EnvironmentDomain = "MG_DOMAIN"
)

var (
	ErrMissingAPIKey = errors.New("Mailgun requires an API key")
	ErrMissingDomain = errors.New("Mailgun requires API domain")
)

type Configuration struct {
	APIKey   string
	Domain   string
	TestMode bool
}

// New creates a Mailgun sending agent.
func New(config Configuration) (mdsend.Sender, error) {
	config.APIKey = strings.TrimSpace(config.APIKey)
	if config.APIKey == "" {
		config.APIKey = strings.TrimSpace(os.Getenv(EnvironmentKey))
	}
	if config.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	config.Domain = strings.TrimSpace(config.Domain)
	if config.Domain == "" {
		config.Domain = strings.TrimSpace(os.Getenv(EnvironmentDomain))
	}
	if config.Domain == "" {
		return nil, ErrMissingDomain
	}
	if !config.TestMode {
		config.TestMode = strings.TrimSpace(os.Getenv("DEBUG")) != ""
	}
	if config.TestMode {
		return mailgunSender{
			MailgunImpl: mailgun.NewMailgun(config.Domain, config.APIKey),
		}.TestMode(), nil
	}
	return mailgunSender{
		MailgunImpl: mailgun.NewMailgun(config.Domain, config.APIKey),
	}, nil
}

type mailgunSender struct {
	*mailgun.MailgunImpl
}

// Send queues one message to one recipient.
func (s mailgunSender) Send(ctx context.Context, d mdsend.Dispatch) (_ string, err error) {
	// message.EnableTestMode()
	// message := s.NewMIMEMessage(d.MIME, d.To)
	// message.EnableNativeSend() // this one fails to deliver mail

	// first returned value is human readable status, second is message ID
	_, id, err := s.MailgunImpl.Send(ctx, s.prepareMessage(d))
	return id, err
}

func (s mailgunSender) TestMode() mdsend.Sender {
	return mailgunTestSender{
		mailgunSender{
			MailgunImpl: s.MailgunImpl,
		},
	}
}

type mailgunTestSender struct {
	mailgunSender
}

func (s mailgunTestSender) Send(ctx context.Context, d mdsend.Dispatch) (_ string, err error) {
	message := s.prepareMessage(d)
	message.EnableTestMode()
	_, id, err := s.MailgunImpl.Send(ctx, message)
	return id, err
}
