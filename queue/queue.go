/*
Package queue stores electronic mail messages into a database.
*/
package queue

import (
	"context"
	"iter"
)

type Queue interface {
	Schedule(context.Context, Message) error
	Cancel(ctx context.Context, ID string) error
	List(context.Context) iter.Seq2[Message, error]
}
