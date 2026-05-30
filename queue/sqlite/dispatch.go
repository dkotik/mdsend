package sqlite

import (
	"context"
	"errors"
	"iter"

	"github.com/dkotik/mdsend"
)

func (q queue) ListDispatches(ctx context.Context, ID string) iter.Seq2[mdsend.Dispatch, error] {
	return func(yield func(mdsend.Dispatch, error) bool) {
		yield(mdsend.Dispatch{}, errors.New("impl"))
	}
}

func (q queue) CompleteDispatch(ctx context.Context, ID string) error {
	return errors.New("impl")
}
