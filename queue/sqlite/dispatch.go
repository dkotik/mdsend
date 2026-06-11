package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/mail"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/oklog/ulid/v2"
	lib "modernc.org/sqlite/lib"
	"zombiezen.com/go/sqlite"
)

func (q queue) CreateDispatch(
	ctx context.Context,
	d mdsend.Dispatch,
) (err error) {
	if err = q.stmtInsertDispatch.Reset(); err != nil {
		return err
	}
	b := &bytes.Buffer{}
	if err = json.NewEncoder(b).Encode(d.Headers); err != nil {
		return err
	}

	q.stmtInsertDispatch.BindText(1, ulid.Make().String())
	q.stmtInsertDispatch.BindText(2, d.LetterID)
	q.stmtInsertDispatch.BindText(3, b.String())
	q.stmtInsertDispatch.BindText(4, d.From.Name)
	q.stmtInsertDispatch.BindText(5, d.From.Address)
	q.stmtInsertDispatch.BindText(6, d.To.Name)
	q.stmtInsertDispatch.BindText(7, d.To.Address)
	q.stmtInsertDispatch.BindText(8, d.Subject)
	q.stmtInsertDispatch.BindText(9, d.Text)
	q.stmtInsertDispatch.BindText(10, d.HTML)
	_, err = q.stmtInsertDispatch.Step()
	switch code := sqlite.ErrCode(err); code {
	case lib.SQLITE_OK:
		return nil
	case lib.SQLITE_CONSTRAINT_UNIQUE:
		return mdsend.ErrDuplicateDispatch
	default:
		return err
	}
}

func (q queue) ListDispatches(ctx context.Context, cursor mdsend.ChildCursor) iter.Seq2[mdsend.Dispatch, error] {
	return func(yield func(mdsend.Dispatch, error) bool) {
		q.DB.SetInterrupt(ctx.Done())
		defer q.DB.SetInterrupt(context.Background().Done())
		limit := cursor.Batch
		var dispatch mdsend.Dispatch
		dispatch.LetterID = cursor.ParentID

		stmt := q.stmtListDispatchesForward
		if limit < 0 {
			limit = -limit
			stmt = q.stmtListDispatchesBackward
		}
		err := stmt.Reset()
		if err != nil {
			yield(dispatch, err)
			return
		}

		stmt.BindText(1, cursor.ParentID)
		if cursor.ItemID == "" {
			stmt.BindText(2, "0")
		} else {
			stmt.BindText(2, cursor.ItemID)
		}
		stmt.BindInt64(3, limit)

		for {
			ok, err := stmt.Step()
			if err != nil {
				yield(dispatch, err)
				return
			}
			if !ok {
				if dispatch.ID != "" {
					// could be more pages
					if err = stmt.Reset(); err != nil {
						yield(dispatch, err)
						return
					}
					stmt.BindText(1, cursor.ParentID)
					stmt.BindText(2, dispatch.ID)
					stmt.BindInt64(3, limit)
					// prevent infinite loop
					dispatch.ID = ""
					continue
				}
				return
			}

			dispatch.ID = stmt.ColumnText(0)
			var headers []mdsend.Header
			if err := json.NewDecoder(stmt.ColumnReader(1)).Decode(&headers); err != nil {
				yield(dispatch, fmt.Errorf("unable to decode headers: %w", err))
				return
			}
			dispatch.Headers = headers
			dispatch.From = mail.Address{
				Name:    stmt.ColumnText(2),
				Address: stmt.ColumnText(3),
			}
			dispatch.To = mail.Address{
				Name:    stmt.ColumnText(4),
				Address: stmt.ColumnText(5),
			}
			dispatch.Subject = stmt.ColumnText(6)
			dispatch.Text = stmt.ColumnText(7)
			dispatch.HTML = stmt.ColumnText(8)

			if !stmt.ColumnIsNull(9) {
				dispatch.SentAt, err = decodeTime(stmt.ColumnText(9))
				if err != nil {
					yield(dispatch, err)
					return
				}
			}
			if !yield(dispatch, nil) {
				return
			}
		}
	}
}

func (q queue) CompleteDispatch(ctx context.Context, ID string) (err error) {

	if err = q.stmtCompleteDispatch.Reset(); err != nil {
		return err
	}
	q.stmtCompleteDispatch.BindText(1, encodeTime(time.Now()))
	q.stmtCompleteDispatch.BindText(2, ID)
	if _, err = q.stmtCompleteDispatch.Step(); err != nil {
		return err
	}
	return nil
}
