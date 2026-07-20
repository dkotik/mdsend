package sparkpost

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"

	sp "github.com/SparkPost/gosparkpost"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/mime"
	"github.com/dkotik/mdsend/queue"
)

const (
	MailerName         = "sparkpost"
	EnvironmentKey     = "SPARKPOST_API_KEY"
	EnvironmentEmailTo = "SPARKPOST_EMAIL_TO"
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
	cfg := &sp.Config{
		BaseUrl:    "https://api.sparkpost.com",
		ApiKey:     config.APIKey,
		ApiVersion: 1,
	}
	var client sp.Client
	err := client.Init(cfg)

	if err != nil {
		return nil, err
	}

	return mailer{
		Client: client,
		Queue:  config.Queue,
		Buffer: bytes.NewBuffer(nil),
	}, nil
}

type mailer struct {
	sp.Client
	Queue  queue.Queue
	Buffer *bytes.Buffer
}

type MimeContent struct {
	EmailRFC822 string `json:"email_rfc822"`
}

// SendMail queues one message to one recipient.
func (s mailer) SendMail(ctx context.Context, m mdsend.Message) (_ string, err error) {
	defer s.Buffer.Reset()
	if err = mime.NewWriter(s.Queue, nil).Write(ctx, s.Buffer, m); err != nil {
		return "", err
	}
	// headers["Idempotency-Key"] = m.SeedKey

	tx := &sp.Transmission{
		Recipients: []string{m.To.Address},
		Content: MimeContent{
			EmailRFC822: s.Buffer.String(),
		},
	}
	id, _, err := s.Client.Send(tx)
	return id, err
}
