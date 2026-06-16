package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
)

func (q sqliteQueue) CreateLetter(
	ctx context.Context,
	l mdsend.Letter,
) (err error) {
	defer q.BindContext(ctx)()
	frontmatter, err := json.Marshal(l.Frontmatter)
	if err != nil {
		return err
	}
	if err = q.stmtInsertLetter.Reset(); err != nil {
		return err
	}

	q.stmtInsertLetter.BindText(1, l.ID)
	q.stmtInsertLetter.BindText(2, string(frontmatter))
	q.stmtInsertLetter.BindText(3, l.Content)
	q.stmtInsertLetter.BindText(4, encodeTime(l.CreatedAt))
	if l.SentAt.IsZero() {
		// q.stmtInsertLetter.BindNull(5)
	} else {
		q.stmtInsertLetter.BindText(5, encodeTime(l.SentAt))
	}
	_, err = q.stmtInsertLetter.Step()
	if err != nil {
		return err
	}
	return err
}

func (q sqliteQueue) RetrieveLetter(ctx context.Context, ID string) (result mdsend.Letter, err error) {
	defer q.BindContext(ctx)()

	if err = q.stmtRetrieveLetter.Reset(); err != nil {
		return result, err
	}
	q.stmtRetrieveLetter.BindText(1, ID)

	for {
		ok, err := q.stmtRetrieveLetter.Step()
		if err != nil {
			return result, err
		}
		if !ok {
			break
		}
		if err := json.NewDecoder(q.stmtRetrieveLetter.ColumnReader(0)).Decode(&result.Frontmatter); err != nil {
			return result, fmt.Errorf("unable to decode frontmatter: %w", err)
		}
		result.Content = q.stmtRetrieveLetter.ColumnText(1)
		if result.CreatedAt, err = decodeTime(q.stmtRetrieveLetter.ColumnText(2)); err != nil {
			return result, err
		}
		if result.SentAt, err = decodeTime(q.stmtRetrieveLetter.ColumnText(3)); err != nil {
			return result, err
		}
	}
	if result.CreatedAt.IsZero() {
		return result, mdsend.ErrLetterNotFound
	}
	result.ID = ID
	return result, nil
}

func (q sqliteQueue) UpdateLetter(ctx context.Context, l mdsend.Letter) (err error) {
	defer q.BindContext(ctx)()

	if err = q.stmtUpdateLetter.Reset(); err != nil {
		return err
	}
	frontmatter, err := json.Marshal(l.Frontmatter)
	if err != nil {
		return err
	}
	q.stmtUpdateLetter.BindText(1, string(frontmatter))
	q.stmtUpdateLetter.BindText(2, l.Content)
	q.stmtUpdateLetter.BindText(3, l.SentAt.Format(time.RFC3339))
	q.stmtUpdateLetter.BindText(4, l.ID)

	for {
		ok, err := q.stmtUpdateLetter.Step()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (q sqliteQueue) MarkLetterAsSent(ctx context.Context, ID string) (ok bool, err error) {
	defer q.BindContext(ctx)()

	if err = q.stmtMarkLetterAsSent.Reset(); err != nil {
		return false, err
	}
	q.stmtMarkLetterAsSent.BindText(1, encodeTime(time.Now()))
	q.stmtMarkLetterAsSent.BindText(2, ID)
	q.stmtMarkLetterAsSent.BindText(3, ID)

	for {
		ok, err := q.stmtMarkLetterAsSent.Step()
		if err != nil {
			return false, err
		}
		if !ok {
			break
		}
	}
	return q.DB.Changes() > 0, nil
}

func (q sqliteQueue) DeleteLetter(ctx context.Context, ID string) (err error) {
	defer q.BindContext(ctx)()

	if err = q.stmtDeleteLetter.Reset(); err != nil {
		return err
	}
	q.stmtDeleteLetter.BindText(1, ID)

	for {
		ok, err := q.stmtDeleteLetter.Step()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}

	// TODO: manual deletion is present, because foreign key CASCADE
	// constraints are not enforced for some reason
	if err = q.stmtDeleteLetterAttachments.Reset(); err != nil {
		return err
	}
	q.stmtDeleteLetterAttachments.BindText(1, ID)

	for {
		ok, err := q.stmtDeleteLetterAttachments.Step()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}

	if err = q.stmtDeleteLetterDispatches.Reset(); err != nil {
		return err
	}
	q.stmtDeleteLetterDispatches.BindText(1, ID)

	for {
		ok, err := q.stmtDeleteLetterDispatches.Step()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (q sqliteQueue) ListLetters(ctx context.Context, cursor queue.Cursor) iter.Seq2[mdsend.Letter, error] {
	return func(yield func(mdsend.Letter, error) bool) {
		defer q.BindContext(ctx)()
		limit := cursor.Batch
		var letter mdsend.Letter
		// lastID := cursor.ItemID

		stmt := q.stmtListLettersForward
		if limit < 0 {
			limit = -limit
			stmt = q.stmtListLettersBackward
		}
		err := stmt.Reset()
		if err != nil {
			yield(letter, err)
			return
		}

		if cursor.ItemID == "" {
			stmt.BindText(1, "0")
		} else {
			stmt.BindText(1, cursor.ItemID)
		}
		stmt.BindInt64(2, limit)
		for {
			ok, err := stmt.Step()
			if err != nil {
				yield(letter, err)
				return
			}
			if !ok {
				if letter.ID != "" {
					// could be more pages
					if err = stmt.Reset(); err != nil {
						yield(letter, err)
						return
					}
					stmt.BindText(1, letter.ID)
					stmt.BindInt64(2, limit)
					// prevent infinite loop
					letter.ID = ""
					continue
				}
				return
			}

			letter.ID = stmt.ColumnText(0)
			if err := json.NewDecoder(stmt.ColumnReader(1)).Decode(&letter.Frontmatter); err != nil {
				yield(letter, fmt.Errorf("unable to decode frontmatter: %w", err))
				return
			}
			letter.Content = stmt.ColumnText(2)
			letter.CreatedAt, err = decodeTime(stmt.ColumnText(3))
			if err != nil {
				yield(letter, err)
				return
			}
			letter.SentAt, err = decodeTime(stmt.ColumnText(4))
			if err != nil {
				yield(letter, err)
				return
			}
			if !yield(letter, nil) {
				return
			}
		}
	}
}
