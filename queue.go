package mdsend

import (
	"context"
	"iter"
)

type Queue interface {
	CreateLetter(context.Context, Letter, []Attachment, []Dispatch) error
	RetrieveLetter(context.Context, string) (Letter, error)
	DeleteLetter(context.Context, string) error
	ListLetters(context.Context) iter.Seq2[Letter, error]
	ListAttachments(context.Context, string) iter.Seq2[Attachment, error]
	ListDispatches(context.Context, string) iter.Seq2[Dispatch, error]
	CompleteDispatch(context.Context, string) error
}
