package awsses

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/mime"
	"github.com/dkotik/mdsend/queue"
)

var _ mdsend.Mailer = (*mailer)(nil)

const (
	MailerName        = "awsses"
	EnvironmentKey    = "AWS_API_KEY"
	EnvironmentSecret = "AWS_API_SECRET"
)

type mailer struct {
	Queue  queue.Queue
	Client *ses.Client
	Buffer *bytes.Buffer
}

// IAM Permissions: The IAM user or role executing this code must have the ses:SendRawEmail
// cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
//
//	if err != nil {
//		log.Fatalf("unable to load SDK config: %v", err)
//	}
//
// sesClient := ses.NewFromConfig(cfg)
// https://docs.aws.amazon.com/ses/latest/dg/send-email-raw.html
func New(q queue.Queue, client *ses.Client) (mdsend.Mailer, error) {
	if q == nil {
		return nil, errors.New("mail queue is required")
	}
	if client == nil {
		return nil, errors.New("SES client is required")
	}
	return mailer{
		Queue:  q,
		Client: client,
		Buffer: &bytes.Buffer{},
	}, nil
}

func (s mailer) SendMail(ctx context.Context, m mdsend.Message) (_ string, err error) {
	s.Buffer.Reset()
	if err = mime.NewWriter(s.Queue, nil).Write(ctx, s.Buffer, m); err != nil {
		return "", err
	}

	sender := m.From.Address
	output, err := s.Client.SendRawEmail(ctx, &ses.SendRawEmailInput{
		Source: &sender,
		Destinations: []string{
			m.To.Address,
		},
		RawMessage: &types.RawMessage{
			Data: s.Buffer.Bytes(),
		},
	})
	if err != nil {
		return "", err
	}
	return *output.MessageId, nil
}

func main() {
	ctx := context.TODO()

	// 1. Define email parameters
	sender := "sender@yourdomain.com"
	recipient := "recipient@example.com"
	// subject := "Hello from Go and AWS SES!"
	// htmlBody := "<h1>Success!</h1><p>This is a raw MIME message sent via AWS SES.</p>"

	// // Mock file data for attachment
	// attachmentData := []byte("This is the content of your attached text file.")
	// attachmentName := "document.txt"

	// 2. Build the MIME message using enmime
	// ....

	// Encode the MIME data into a raw byte buffer
	var rawBuffer bytes.Buffer

	// 3. Load AWS Configuration (reads default credentials/region)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	sesClient := ses.NewFromConfig(cfg)

	// 4. Construct the SendRawEmail input
	// Note: The AWS SDK automatically handles the outer base64 encoding for RawMessage.Data
	input := &ses.SendRawEmailInput{
		Source: &sender,
		Destinations: []string{
			recipient,
		},
		RawMessage: &types.RawMessage{
			Data: rawBuffer.Bytes(),
		},
	}

	// 5. Send the email
	output, err := sesClient.SendRawEmail(ctx, input)
	if err != nil {
		log.Fatalf("failed to send raw email via SES: %v", err)
	}

	fmt.Printf("Email sent successfully! Message ID: %s\n", *output.MessageId)
}
