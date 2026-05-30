package mdsend

import (
	"context"
	"io"
	"net/mail"
	"time"
)

type Message interface {
	To() []string
	CC() []string
	BCC() []string
	Subject() string
	MIMEBody() io.ReadCloser
}

type Dispatch struct {
	ID        string
	LetterID  string
	Recipient map[string]any
	SentAt    time.Time
}

func (d Dispatch) GetAddress() (mail.Address, error) {
	return newAddressFromMap(d.Recipient)
}

type Sender interface {
	Send(context.Context, Dispatch) error
}
