package sqlite

import (
	"context"
	"errors"
)

func (q sqliteQueue) Cancel(ctx context.Context, ID string) error {
	return errors.New("implement")
}
