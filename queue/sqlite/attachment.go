package sqlite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"

	"github.com/dkotik/mdsend"
	"github.com/oklog/ulid/v2"
	lib "modernc.org/sqlite/lib"
	"zombiezen.com/go/sqlite"
)

func (q sqliteQueue) CreateAttachment(
	ctx context.Context,
	a mdsend.Attachment,
) (err error) {
	if err = q.stmtInsertAttachment.Reset(); err != nil {
		return err
	}

	q.stmtInsertAttachment.BindText(1, ulid.Make().String())
	q.stmtInsertAttachment.BindText(2, a.LetterID)
	q.stmtInsertAttachment.BindText(3, a.Name)
	q.stmtInsertAttachment.BindText(4, a.Hash)
	q.stmtInsertAttachment.BindText(5, a.ContentID)
	q.stmtInsertAttachment.BindText(6, a.ContentType)
	q.stmtInsertAttachment.BindBytes(7, a.Content)

	ok := false
	for {
		ok, err = q.stmtInsertAttachment.Step()
		if err != nil {
			switch code := sqlite.ErrCode(err); code {
			case lib.SQLITE_OK:
				return nil
			case lib.SQLITE_CONSTRAINT_UNIQUE:
				return fmt.Errorf("error putting %q (#%s) into queue: %w", a.Name, a.Hash, mdsend.ErrDuplicateAttachment)
			default:
				return err
			}
		}
		if !ok {
			break
		}
	}
	return nil
}

func (q sqliteQueue) ListAttachments(ctx context.Context, letterID string) iter.Seq2[mdsend.Attachment, error] {
	return func(yield func(mdsend.Attachment, error) bool) {
		defer q.BindContext(ctx)()
		stmt := q.stmtListAttachments
		var err error
		defer func() {
			err = errors.Join(err, stmt.Reset())
		}()

		stmt.BindText(1, letterID)

		for {
			ok, err := stmt.Step()
			if err != nil {
				yield(mdsend.Attachment{}, err)
				return
			}
			if !ok {
				break
			}
			b := &bytes.Buffer{}
			if _, err := io.Copy(b, stmt.ColumnReader(4)); err != nil {
				yield(mdsend.Attachment{}, err)
				return
			}
			if !yield(mdsend.Attachment{
				LetterID:    letterID,
				Name:        stmt.ColumnText(0),
				Hash:        stmt.ColumnText(1),
				ContentID:   stmt.ColumnText(2),
				ContentType: stmt.ColumnText(3),
				Content:     b.Bytes(),
			}, err) {
				return
			}
		}
	}
}
