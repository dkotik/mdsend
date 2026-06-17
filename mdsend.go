package mdsend

import (
	"context"
)

const Version = "dev"

type Mailer interface {
	SendMail(context.Context, Message) (string, error)
}
