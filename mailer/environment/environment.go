package environment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/mailer/awsses"
	"github.com/dkotik/mdsend/mailer/mailgun"
	"github.com/dkotik/mdsend/mailer/resend"
	"github.com/dkotik/mdsend/mailer/smtp"
	"github.com/dkotik/mdsend/queue"
)

var ErrNoCredentials = errors.New("there are no environment credentials for any mail driver")

func New(
	ctx context.Context,
	q queue.Queue,
	logger *slog.Logger,
	mailerNamePriority ...string,
) (mdsend.Mailer, error) {
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
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				logger.Debug("failed to load AWS SES credentials from environment",
					slog.String("mailer", mailerName),
					slog.Any("error", err),
				)
				continue
			}
			// If config is empty, anonymous credentials will be used
			// by default. Attempt to retrieve credentials to confirm
			// that the config is valid.
			_, err = cfg.Credentials.Retrieve(ctx)
			if err == nil {
				return awsses.New(q, sesv2.NewFromConfig(cfg))
			}
			logger.Debug("attempted to activate AWS SES credentials from environment, but failed to obtain them",
				slog.String("mailer", mailerName),
				slog.Any("error", err),
			)
		default:
			return nil, fmt.Errorf("unsupporter mailer driver: <%s>", mailerName)
		}
		logger.Debug("valid mailer credentials for one driver, in order of priority, were not discovered in the environment, moving on to the next driver", slog.String("mailer", mailerName))
	}
	return nil, ErrNoCredentials
}
