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

func (q sqliteQueue) CreateMessage(
	ctx context.Context,
	d mdsend.Message,
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
	if !d.ScheduleAfter.IsZero() {
		q.stmtInsertMessage.BindText(11, d.ScheduleAfter.Format(time.RFC3339))
	}
	_, err = q.stmtInsertMessage.Step()
	switch code := sqlite.ErrCode(err); code {
	case lib.SQLITE_OK:
		return nil
	case lib.SQLITE_CONSTRAINT_PRIMARYKEY, lib.SQLITE_CONSTRAINT_UNIQUE:
		// same id or same combination of letter_id and recipient address
		return mdsend.ErrDuplicateMessage
	default:
		return err
	}
}

func (q sqliteQueue) selectResetBindListMessagesStatement(letterID, ID string, batch int64) (stmt *sqlite.Stmt, err error) {
	if batch < 0 {
		if ID == "" {
			stmt = q.stmtListMessagesBackwardHead
			if err = stmt.Reset(); err != nil {
				return stmt, err
			}
			stmt.BindText(1, letterID)
			stmt.BindInt64(2, -batch)
			return stmt, nil
		}
		stmt = q.stmtListMessagesBackward
		if err = stmt.Reset(); err != nil {
			return stmt, err
		}
		stmt.BindText(1, letterID)
		stmt.BindText(2, ID)
		stmt.BindInt64(3, -batch)
		return stmt, nil
	}
	stmt = q.stmtLisMessagesForward
	if err = stmt.Reset(); err != nil {
		return stmt, err
	}
	stmt.BindText(1, letterID)
	stmt.BindText(2, ID)
	stmt.BindInt64(3, batch)
	return stmt, nil
}

func (q sqliteQueue) ListMessages(ctx context.Context, cursor queue.ChildCursor) iter.Seq2[mdsend.Message, error] {
	return func(yield func(mdsend.Message, error) bool) {
		defer q.BindContext(ctx)()
		var m mdsend.Message
		m.LetterID = cursor.ParentID // propagate parent to all returned rows
		stmt, err := q.selectResetBindListMessagesStatement(cursor.ParentID, cursor.ItemID, cursor.Batch)
		if err != nil {
			yield(m, err)
			return
		}

		for {
			ok, err := stmt.Step()
			if err != nil {
				yield(m, err)
				return
			}
			if !ok {
				if m.ID != "" {
					stmt, err = q.selectResetBindListMessagesStatement(cursor.ParentID, m.ID, cursor.Batch)
					if err != nil {
						yield(m, err)
						return
					}
					// prevent infinite loop
					m.ID = ""
					// goto nextPage
					continue
					// return
				}
				return
			}

			m.ID = stmt.ColumnText(0)
			var headers []mdsend.Header
			if err := json.NewDecoder(stmt.ColumnReader(1)).Decode(&headers); err != nil {
				yield(m, fmt.Errorf("unable to decode headers: %w", err))
				return
			}
			m.Headers = headers
			m.From = mail.Address{
				Name:    stmt.ColumnText(2),
				Address: stmt.ColumnText(3),
			}
			m.To = mail.Address{
				Name:    stmt.ColumnText(4),
				Address: stmt.ColumnText(5),
			}
			m.Subject = stmt.ColumnText(6)
			m.Text = stmt.ColumnText(7)
			m.HTML = stmt.ColumnText(8)

			if !stmt.ColumnIsNull(9) {
				m.ScheduleAfter, err = decodeTime(stmt.ColumnText(9))
				if err != nil {
					yield(m, err)
					return
				}
			}
			if !stmt.ColumnIsNull(10) {
				m.ScheduledAt, err = decodeTime(stmt.ColumnText(10))
				if err != nil {
					yield(m, err)
					return
				}
			}
			if !stmt.ColumnIsNull(11) {
				m.SentAt, err = decodeTime(stmt.ColumnText(11))
				if err != nil {
					yield(m, err)
					return
				}
			}
			if !yield(m, nil) {
				return
			}
		}
	}
}

func (q sqliteQueue) MarkMessagesAsScheduled(ctx context.Context, letterID string, IDs ...string) (err error) {
	ids, err := json.Marshal(IDs)
	if err != nil {
		return err
	}
	defer q.BindContext(ctx)()
	// TODO: letterID should also be bound
	if err = q.stmtMarkMessagesAsQueued.Reset(); err != nil {
		return err
	}
	q.stmtMarkMessagesAsQueued.BindText(1, encodeTime(time.Now()))
	q.stmtMarkMessagesAsQueued.BindBytes(2, ids)
	for {
		ok, err := q.stmtMarkMessagesAsQueued.Step()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
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
