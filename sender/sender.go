package sender

import (
	"context"

	"github.com/dkotik/mdsend"
)

type void struct{}

func NewVoid() mdsend.Sender {
	return void{}
}

func (v void) Send(ctx context.Context, msg mdsend.Dispatch) (string, error) {
	return msg.ID, nil
}
