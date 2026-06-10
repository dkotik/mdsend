package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"github.com/dkotik/mdsend"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func (q queue) CreateLetter(
	ctx context.Context,
	l mdsend.Letter,
	attachments []mdsend.Attachment,
	dispatches []mdsend.Dispatch,
) (err error) {
	frontmatter, err := json.Marshal(l.Frontmatter)
	if err != nil {
		return err
	}
	if err = sqlitex.Execute(q.DB, q.stmtInsertLetter, &sqlitex.ExecOptions{
		Args: []any{
			l.ID,
			frontmatter,
			l.Content,
			encodeTime(l.CreatedAt),
			encodeTime(l.SentAt),
		},
	}); err != nil {
		return err
	}
	return nil
}

func (q queue) RetrieveLetter(ctx context.Context, ID string) (result mdsend.Letter, err error) {
	if err = sqlitex.Execute(q.DB, q.stmtRetrieveLetter, &sqlitex.ExecOptions{
		Args: []any{ID},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if err := json.NewDecoder(stmt.ColumnReader(0)).Decode(&result.Frontmatter); err != nil {
				return fmt.Errorf("unable to decode frontmatter: %w", err)
			}
			result.Content = stmt.ColumnText(1)
			if result.CreatedAt, err = decodeTime(stmt.ColumnText(2)); err != nil {
				return err
			}
			if result.SentAt, err = decodeTime(stmt.ColumnText(3)); err != nil {
				return err
			}
			return nil
		},
	}); err != nil {
		return result, err
	}
	result.ID = ID
	return result, nil
}

func (q queue) DeleteLetter(ctx context.Context, ID string) error {
	return sqlitex.Execute(q.DB, q.stmtDeleteLetter, &sqlitex.ExecOptions{
		Args: []any{ID},
	})
}

func (q queue) ListLetters(ctx context.Context) iter.Seq2[mdsend.Letter, error] {
	return func(yield func(mdsend.Letter, error) bool) {
		stmt, err := q.DB.Prepare(q.stmtListLetters)
		if err != nil {
			yield(mdsend.Letter{}, err)
			return
		}
		defer stmt.Finalize()
		for {
			ok, err := stmt.Step()
			if err != nil {
				yield(mdsend.Letter{}, err)
				return
			}
			if !ok {
				return
			}
			var l mdsend.Letter
			l.ID = stmt.ColumnText(0)
			if err := json.NewDecoder(stmt.ColumnReader(1)).Decode(&l.Frontmatter); err != nil {
				yield(mdsend.Letter{}, fmt.Errorf("unable to decode frontmatter: %w", err))
				return
			}
			l.Content = stmt.ColumnText(2)
			l.CreatedAt, err = decodeTime(stmt.ColumnText(3))
			if err != nil {
				yield(mdsend.Letter{}, err)
				return
			}
			l.SentAt, err = decodeTime(stmt.ColumnText(4))
			if err != nil {
				yield(mdsend.Letter{}, err)
				return
			}
			if !yield(l, nil) {
				return
			}
		}
	}
}
