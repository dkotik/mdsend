package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"github.com/dkotik/mdsend"
)

func (q queue) CreateLetter(
	ctx context.Context,
	l mdsend.Letter,
	attachments []mdsend.Attachment,
	dispatches []mdsend.Dispatch,
) (err error) {
	q.DB.SetInterrupt(ctx.Done())
	defer q.DB.SetInterrupt(context.Background().Done())
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
	q.stmtInsertLetter.BindText(5, encodeTime(l.SentAt))
	_, err = q.stmtInsertLetter.Step()
	if err != nil {
		return err
	}
	for _, a := range attachments {
		if err = q.CreateAttachment(ctx, a); err != nil {
			return err
		}
	}

	return err
}

func (q queue) RetrieveLetter(ctx context.Context, ID string) (result mdsend.Letter, err error) {
	q.DB.SetInterrupt(ctx.Done())
	defer q.DB.SetInterrupt(context.Background().Done())

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
	result.ID = ID
	return result, nil
}

func (q queue) DeleteLetter(ctx context.Context, ID string) (err error) {
	q.DB.SetInterrupt(ctx.Done())
	defer q.DB.SetInterrupt(context.Background().Done())

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

func (q queue) ListLetters(ctx context.Context, cursor mdsend.Cursor) iter.Seq2[mdsend.Letter, error] {
	return func(yield func(mdsend.Letter, error) bool) {
		q.DB.SetInterrupt(ctx.Done())
		defer q.DB.SetInterrupt(context.Background().Done())
		limit := cursor.Batch
		var letter mdsend.Letter
		// lastID := cursor.ItemID

		stmt := q.stmtListLettersForward
		if limit < 0 {
			limit = -limit
			stmt = q.stmtListLettersBackward
		}
		err := stmt.Reset()
		if cursor.ItemID == "" {
			stmt.BindText(1, "0")
		} else {
			stmt.BindText(1, cursor.ItemID)
		}
		stmt.BindInt64(2, limit)
		if err != nil {
			yield(letter, err)
			return
		}
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
