package resend

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
	"github.com/resendlabs/resend-go"
)

const (
	MailerName         = "resend"
	EnvironmentKey     = "RESEND_API_KEY"
	EnvironmentEmailTo = "RESEND_EMAIL_TO"
)

var (
	ErrMissingAPIKey = errors.New("Resend requires an API key")
)

type Configuration struct {
	Queue    queue.Queue
	APIKey   string
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
	if !config.TestMode {
		config.TestMode = strings.TrimSpace(os.Getenv("DEBUG")) != ""
	}
	if config.TestMode {
		return nil, errors.New("test mode is not yet implemented for Resend")
	}
	return sender{
		Client: resend.NewClient(config.APIKey),
		Queue:  config.Queue,
		Buffer: bytes.NewBuffer(nil),
	}, nil
}

type sender struct {
	*resend.Client
	Queue  queue.Queue
	Buffer *bytes.Buffer
}

// SendMail queues one message to one recipient.
func (s sender) SendMail(ctx context.Context, d mdsend.Message) (_ string, err error) {
	// TODO: this could be cached by wrapping the queue?
	attachments := make([]resend.Attachment, 0)
	for attachment, err := range s.Queue.ListAttachments(ctx, d.LetterID) {
		if err != nil {
			return "", err
		}
		attachments = append(attachments, resend.Attachment{
			Filename: attachment.Name,
			Content:  string(attachment.Content),
		})
	}

	request := &resend.SendEmailRequest{
		From:    d.From.String(),
		To:      []string{d.To.String()},
		Subject: d.Subject,
		// Cc:          cc,
		// Bcc:         bcc,
		Html:        d.HTML,
		Text:        d.Text,
		Attachments: attachments,
	}

	response, err := s.Client.Emails.Send(request)
	if err != nil {
		return "", err
	}
	return response.Id, err
}
