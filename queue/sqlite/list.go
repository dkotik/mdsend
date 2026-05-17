package sqlite

import (
	"context"
	"iter"

	q "github.com/dkotik/mdsend/queue"
)

func (q queue) List(ctx context.Context) iter.Seq2[q.Message, error] {
	return nil
}
