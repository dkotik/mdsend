package environment

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/mailer/awsses"
	"github.com/dkotik/mdsend/mailer/mailgun"
	"github.com/dkotik/mdsend/mailer/resend"
	"github.com/dkotik/mdsend/mailer/smtp"
	"github.com/dkotik/mdsend/queue"
)

func New(q queue.Queue, mailerNamePriority ...string) (mdsend.Mailer, error) {
	mailerNamePriority = append(
		mailerNamePriority,
		mailgun.MailerName,
		resend.MailerName,
		smtp.MailerName,
		awsses.MailerName,
	)
	for _, mailerName := range mailerNamePriority {
		switch mailerName {
		case mailgun.MailerName:
			apiKey := strings.TrimSpace(os.Getenv(mailgun.EnvironmentKey))
			if apiKey != "" {
				return mailgun.New(mailgun.Configuration{
					Queue:  q,
					APIKey: apiKey,
				})
			}
		case resend.MailerName:
			apiKey := strings.TrimSpace(os.Getenv(resend.EnvironmentKey))
			if apiKey != "" {
				return resend.New(resend.Configuration{
					Queue:  q,
					APIKey: apiKey,
				})
			}
		case smtp.MailerName:
			userName := strings.TrimSpace(os.Getenv(smtp.EnvironmentUsername))
			if userName != "" {
				return smtp.New(smtp.Configuration{
					Queue: q,
					// Username: userName,
				})
			}
		case awsses.MailerName:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
			defer cancel()
			cfg, _ := config.LoadDefaultConfig(ctx)
			if cfg.Credentials != nil {
				return awsses.New(q, sesv2.NewFromConfig(cfg))
			}
		default:
			return nil, fmt.Errorf("unsupporter mailer driver: <%s>", mailerName)
		}
	}
	return nil, errors.New("there are no environment credentials for any mail driver")
}
