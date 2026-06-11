package mdsend

import (
	"context"
	"iter"
)

// Cursor holds the position of an item in a list from which an iterator
// can retrieve items sequentially. If [Cursor.ItemID] is empty, the iterator
// starts from the first batch. The item with the same ID is
// always skipped, so the iterator will start from the next item.
//
// [Cursor.Batch] sets the maximum number of items to retrieve in one
// repository paging operation. A negative batch value iterates items
// in descending order from the [Cursor.ItemID].
//
// The iterator loads additional batches as needed as long as the range
// of items to retrieve is not exhausted.
//
// Context cancellation will stop the iterator at the end
// of the current batch.
type Cursor struct {
	ItemID string
	Batch  int64
}

type ChildCursor struct {
	ParentID string
	Cursor
}

type Queue interface {
	CreateLetter(context.Context, Letter, []Attachment, []Dispatch) error
	RetrieveLetter(context.Context, string) (Letter, error)
	DeleteLetter(context.Context, string) error
	CompleteDispatch(context.Context, string) error

	ListLetters(context.Context, Cursor) iter.Seq2[Letter, error]
	ListDispatches(context.Context, ChildCursor) iter.Seq2[Dispatch, error]
	ListAttachments(context.Context, string) iter.Seq2[Attachment, error]
}
