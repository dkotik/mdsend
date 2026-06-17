package mdsend

import (
	"context"
)

const Version = "dev"

type Sender interface {
	Send(context.Context, Message) (string, error)
}
