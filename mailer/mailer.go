package mailer

import (
	"context"

	"github.com/dkotik/mdsend"
)

type void struct{}

func NewVoid() mdsend.Mailer {
	return void{}
}

func (v void) SendMail(ctx context.Context, msg mdsend.Message) (string, error) {
	return msg.ID, nil
}
