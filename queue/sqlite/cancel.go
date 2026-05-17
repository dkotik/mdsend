package sqlite

import (
	"context"
	"errors"
)

func (q queue) Cancel(ctx context.Context, ID string) error {
	return errors.New("implement")
}
