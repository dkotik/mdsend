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
	"github.com/dkotik/mdsend/queue"
	lib "modernc.org/sqlite/lib"
	"zombiezen.com/go/sqlite"
)

func (q sqliteQueue) CreateDispatch(
	ctx context.Context,
	d mdsend.Dispatch,
) (err error) {
	if err = q.stmtInsertMessage.Reset(); err != nil {
		return err
	}
	b := &bytes.Buffer{}
	if err = json.NewEncoder(b).Encode(d.Headers); err != nil {
		return err
	}

	q.stmtInsertMessage.BindText(1, d.ID)
	q.stmtInsertMessage.BindText(2, d.LetterID)
	q.stmtInsertMessage.BindText(3, b.String())
	q.stmtInsertMessage.BindText(4, d.From.Name)
	q.stmtInsertMessage.BindText(5, d.From.Address)
	q.stmtInsertMessage.BindText(6, d.To.Name)
	q.stmtInsertMessage.BindText(7, d.To.Address)
	q.stmtInsertMessage.BindText(8, d.Subject)
	q.stmtInsertMessage.BindText(9, d.Text)
	q.stmtInsertMessage.BindText(10, d.HTML)
	_, err = q.stmtInsertMessage.Step()
	switch code := sqlite.ErrCode(err); code {
	case lib.SQLITE_OK:
		return nil
	case lib.SQLITE_CONSTRAINT_UNIQUE:
		return mdsend.ErrDuplicateDispatch
	default:
		return err
	}
}

func (q sqliteQueue) ListDispatches(ctx context.Context, cursor queue.ChildCursor) iter.Seq2[mdsend.Dispatch, error] {
	return func(yield func(mdsend.Dispatch, error) bool) {
		defer q.BindContext(ctx)()
		limit := cursor.Batch
		var dispatch mdsend.Dispatch
		dispatch.LetterID = cursor.ParentID

		stmt := q.stmtListDispatchesForward
		if limit < 0 {
			limit = -limit
			stmt = q.stmtListDispatchesBackward
		}

		// nextPage:
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
					// goto nextPage
					continue
					// return
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

func (q sqliteQueue) MarkMessageAsQueued(ctx context.Context, ID string) (ok bool, err error) {
	defer q.BindContext(ctx)()
	if err = q.stmtMarkMessageAsQueued.Reset(); err != nil {
		return false, err
	}
	q.stmtMarkMessageAsQueued.BindText(1, encodeTime(time.Now()))
	q.stmtMarkMessageAsQueued.BindText(2, ID)
	for {
		ok, err := q.stmtMarkMessageAsQueued.Step()
		if err != nil {
			return false, err
		}
		if !ok {
			break
		}
	}
	return q.DB.Changes() > 0, nil
}

func (q sqliteQueue) MarkMessageAsSent(ctx context.Context, ID string) (ok bool, err error) {
	defer q.BindContext(ctx)()
	if err = q.stmtMarkMessageAsSent.Reset(); err != nil {
		return false, err
	}
	q.stmtMarkMessageAsSent.BindText(1, encodeTime(time.Now()))
	q.stmtMarkMessageAsSent.BindText(2, ID)
	for {
		ok, err := q.stmtMarkMessageAsSent.Step()
		if err != nil {
			return false, err
		}
		if !ok {
			break
		}
	}
	return q.DB.Changes() > 0, nil
}
