package mailgun

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
	mailgun "github.com/mailgun/mailgun-go/v4"
)

const (
	EnvironmentKey     = "MG_API_KEY"
	EnvironmentDomain  = "MG_DOMAIN"
	EnvironmentEmailTo = "MG_EMAIL_TO"
)

var (
	ErrMissingAPIKey = errors.New("Mailgun requires an API key")
	ErrMissingDomain = errors.New("Mailgun requires API domain")
)

type Configuration struct {
	Queue    queue.Queue
	APIKey   string
	Domain   string
	TestMode bool
}

// New creates a Mailgun sending agent.
func New(config Configuration) (mdsend.Mailer, error) {
	if config.Queue == nil {
		return nil, errors.New("message queue is required for attachments: nil")
	}
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
			Buffer:      bytes.NewBuffer(nil),
		}.TestMode(), nil
	}
	return mailgunSender{
		MailgunImpl: mailgun.NewMailgun(config.Domain, config.APIKey),
		Queue:       config.Queue,
		Buffer:      bytes.NewBuffer(nil),
	}, nil
}

type mailgunSender struct {
	*mailgun.MailgunImpl
	Queue  queue.Queue
	Buffer *bytes.Buffer
}

// SendMail queues one message to one recipient.
func (s mailgunSender) SendMail(ctx context.Context, d mdsend.Message) (_ string, err error) {
	// first returned value is human readable status, second is message ID
	msg, err := s.prepareMessage(ctx, d)
	if err != nil {
		return "", err
	}
	_, id, err := s.MailgunImpl.Send(ctx, msg)
	return id, err
}

func (s mailgunSender) TestMode() mdsend.Mailer {
	return mailgunTestSender{
		mailgunSender{
			MailgunImpl: s.MailgunImpl,
			Buffer:      s.Buffer,
		},
	}
}

type mailgunTestSender struct {
	mailgunSender
}

func (s mailgunTestSender) SendMail(ctx context.Context, d mdsend.Message) (_ string, err error) {
	message, err := s.prepareMessage(ctx, d)
	defer s.Buffer.Reset()
	if err != nil {
		return "", err
	}
	message.EnableTestMode()
	_, id, err := s.MailgunImpl.Send(ctx, message)
	return id, err
}
