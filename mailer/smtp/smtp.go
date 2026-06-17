/*
Package smtp sends MIME-formatted emails via SMTP.

Credentials are provided by most electronic mail box hosting services. The common ones:

  - Gmail App Passwords: <https://myaccount.google.com/apppasswords>
    Requires two-factor authentication enabled on your Google account.
*/
package smtp

import (
	"bytes"
	"errors"
	"net/smtp"
	"os"
	"strings"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
)

const (
	EnvironmentServer   = "SMTP_SERVER"
	EnvironmentPort     = "SMTP_PORT"
	EnvironmentUsername = "SMTP_USERNAME"
	EnvironmentPassword = "SMTP_PASSWORD"
	EnvironmentTestTo   = "SMTP_TEST_TO"
)

type Configuration struct {
	Server         string
	Port           string
	Queue          queue.Queue
	Authentication smtp.Auth
}

func (c Configuration) withDefaults() (_ Configuration, err error) {
	c.Server = strings.TrimSpace(c.Server)
	c.Port = strings.TrimSpace(c.Port)
	if c.Server == "" {
		c.Server = strings.TrimSpace(os.Getenv(EnvironmentServer))
	}
	if c.Port == "" {
		c.Port = strings.TrimSpace(os.Getenv(EnvironmentPort))
	}
	if c.Server == "" {
		c.Server = "smtp.gmail.com"
	}
	if c.Port == "" {
		c.Port = "587" // modern standard
	}
	if c.Authentication == nil {
		c.Authentication, err = LoginAuth(os.Getenv(EnvironmentUsername), os.Getenv(EnvironmentPassword))
		if err != nil {
			return c, err
		}
	}
	return c, nil
}

func New(config Configuration) (_ mdsend.Mailer, err error) {
	config, err = config.withDefaults()
	if err != nil {
		return nil, err
	}
	if config.Queue == nil {
		return nil, errors.New("queue is nil")
	}
	// if config.Authentication == nil {
	// 	return nil, errors.New("authentication is nil")
	// }
	return senderSMTP{
		Buffer:         bytes.NewBuffer(nil),
		Queue:          config.Queue,
		Authentication: config.Authentication,
		Connection:     config.Server + ":" + config.Port,
	}, nil
}

type senderSMTP struct {
	Buffer         *bytes.Buffer
	Queue          queue.Queue
	Authentication smtp.Auth
	Connection     string
}
