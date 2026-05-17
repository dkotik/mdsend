package sqlite

import (
	"context"
	"errors"

	q "github.com/dkotik/mdsend/queue"
)

func (q queue) Schedule(ctx context.Context, message q.Message) error {
	return errors.New("implement")
}
