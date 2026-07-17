package awsses

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/mime"
	"github.com/dkotik/mdsend/queue"
)

var _ mdsend.Mailer = (*mailer)(nil)

const (
	MailerName        = "awsses"
	EnvironmentKey    = "AWS_API_KEY"
	EnvironmentSecret = "AWS_API_SECRET"

	TestEmailAddress = "success@simulator.amazonses.com"
)

type mailer struct {
	Queue  queue.Queue
	Client *sesv2.Client
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
func New(q queue.Queue, client *sesv2.Client) (mdsend.Mailer, error) {
	if q == nil {
		return nil, errors.New("mail queue is required")
	}
	if client == nil {
		return nil, errors.New("SES client is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
	defer cancel()
	sender := TestEmailAddress
	_, err := client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &sender,
		Destination: &types.Destination{
			ToAddresses: []string{
				sender,
			},
		},
		Content: &types.EmailContent{
			Raw: &types.RawMessage{
				Data: []byte("<html>Empty test</html>"),
			},
		},
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			// AccessDenied means the client lacks ses:SendEmail permissions
			if apiErr.ErrorCode() == "AccessDenied" {
				return nil, fmt.Errorf("permission check failed: %v; does this AWS account have ses:SendRawEmail permission?", apiErr.ErrorMessage())
			}
			if apiErr.ErrorCode() == "MessageRejected" {
				// MessageRejected is expected here because it is a dry run
			}
		} else {
			return nil, err
		}
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
	output, err := s.Client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &sender,
		Destination: &types.Destination{
			ToAddresses: []string{
				m.To.Address,
			},
		},
		Content: &types.EmailContent{
			Raw: &types.RawMessage{
				Data: s.Buffer.Bytes(),
			},
		},
	})
	if err != nil {
		return "", err
	}
	if output.MessageId == nil {
		return "", nil
	}
	return *output.MessageId, nil
}
