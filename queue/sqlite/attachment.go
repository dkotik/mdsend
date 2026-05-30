package sqlite

import (
	"context"
	"errors"
	"iter"

	"github.com/dkotik/mdsend"
)

func (q queue) ListAttachments(ctx context.Context, ID string) iter.Seq2[mdsend.Attachment, error] {
	return func(yield func(mdsend.Attachment, error) bool) {
		yield(mdsend.Attachment{}, errors.New("impl"))
	}
}
