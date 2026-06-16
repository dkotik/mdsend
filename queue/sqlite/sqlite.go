package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dkotik/mdsend/queue"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type transaction func(*error)

func (t transaction) Close(err *error) {
	t(err)
}

type sqliteQueue struct {
	DB *sqlite.Conn

	stmtInsertLetter            *sqlite.Stmt
	stmtInsertMessage           *sqlite.Stmt
	stmtInsertAttachment        *sqlite.Stmt
	stmtRetrieveLetter          *sqlite.Stmt
	stmtUpdateLetter            *sqlite.Stmt
	stmtMarkLetterAsSent        *sqlite.Stmt
	stmtDeleteLetter            *sqlite.Stmt
	stmtDeleteLetterAttachments *sqlite.Stmt
	stmtDeleteLetterDispatches  *sqlite.Stmt
	stmtListLettersForward      *sqlite.Stmt
	stmtListLettersBackward     *sqlite.Stmt
	stmtListAttachments         *sqlite.Stmt
	stmtListDispatchesForward   *sqlite.Stmt
	stmtListDispatchesBackward  *sqlite.Stmt
	stmtMarkMessageAsQueued     *sqlite.Stmt
	stmtMarkMessageAsSent       *sqlite.Stmt
}

// New creates an SQLite3 queue at the location.
func New(conn *sqlite.Conn, prefix string) (_ queue.Queue, err error) {
	if prefix == "" {
		prefix = "mdsend_"
	}
	lettersTable := escapeIdentifier(prefix + "letters")
	attachmentsTable := escapeIdentifier(prefix + "attachments")
	messagesTable := escapeIdentifier(prefix + "messages")

	createLettersTable := true
	createAttachmentsTable := true
	createMessagesTable := true

	if err = sqlitex.ExecuteScript(conn, `
		PRAGMA foreign_keys = ON;

		SELECT name FROM sqlite_master WHERE type='table' AND name LIKE '`+prefix+`%';
	`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			switch stmt.ColumnText(0) {
			case lettersTable:
				createLettersTable = false
			case attachmentsTable:
				createAttachmentsTable = false
			case messagesTable:
				createMessagesTable = false
			}
			return nil
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to list tables: %w", err)
	}

	if createLettersTable || createAttachmentsTable || createMessagesTable {
		if err = sqlitex.ExecScript(
			conn,
			`
		CREATE TABLE IF NOT EXISTS `+lettersTable+` (
			id text PRIMARY KEY,
			frontmatter text NOT NULL,
			content text NOT NULL,
			created_at text NOT NULL,
			queue_after text,
			sent_at text
		) STRICT;

		CREATE INDEX IF NOT EXISTS index_name
			ON `+lettersTable+` (id);

		CREATE TABLE IF NOT EXISTS `+attachmentsTable+` (
			id text PRIMARY KEY,
			letter_id text NOT NULL,
			name text NOT NULL,
			content_hash text NOT NULL,
			content_id text,
			content_type text NOT NULL,
			content blob NOT NULL,

			FOREIGN KEY (letter_id) REFERENCES `+lettersTable+`(id)
	   		ON DELETE CASCADE
	   		ON UPDATE CASCADE,
			UNIQUE (letter_id, content_hash)
		) STRICT;

		CREATE TABLE IF NOT EXISTS `+messagesTable+` (
			id text PRIMARY KEY,
			letter_id text NOT NULL REFERENCES `+lettersTable+`(id) ON DELETE CASCADE,
			headers text NOT NULL,
			from_name text NOT NULL,
			from_email text NOT NULL,
			to_name text NOT NULL,
			to_email text NOT NULL,
			subject text NOT NULL,
			message_text text NOT NULL,
			message_html text NOT NULL,
			queue_after text,
			queued_at text,
			sent_at text,

			UNIQUE (letter_id, to_email)
		) STRICT;
		`,
		); err != nil {
			return nil, fmt.Errorf("unable to create tables: %w", err)
		}
	}
	q := sqliteQueue{
		DB: conn,
	}

	if q.stmtInsertLetter, err = conn.Prepare(`INSERT INTO ` + lettersTable + `(id, frontmatter, content, created_at, sent_at) VALUES(?,?,?,?,?)`); err != nil {
		return nil, fmt.Errorf("unable to prepare insert letter statement: %w", err)
	}
	if q.stmtInsertMessage, err = conn.Prepare(`INSERT INTO ` + messagesTable + ` (id, letter_id, headers, from_name, from_email, to_name, to_email, subject, message_text, message_html) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`); err != nil {
		return nil, fmt.Errorf("unable to prepare insert dispatch statement: %w", err)
	}
	if q.stmtInsertAttachment, err = q.DB.Prepare(`INSERT INTO ` + attachmentsTable + ` (id, letter_id, name, content_hash, content_id, content_type, content) VALUES (?, ?, ?, ?, ?, ?, ?)`); err != nil {
		return nil, fmt.Errorf("unable to prepare insert attachment statement: %w", err)
	}
	if q.stmtRetrieveLetter, err = conn.Prepare(`SELECT frontmatter, content, created_at, sent_at FROM ` + lettersTable + ` WHERE id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare retrieve letter statement: %w", err)
	}
	if q.stmtUpdateLetter, err = conn.Prepare(`UPDATE ` + lettersTable + ` SET frontmatter=?, content=?, sent_at=? WHERE id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare update letter statement: %w", err)
	}
	// AND sent_at IS NULL AND NOT EXISTS (SELECT 1 FROM ` + dispatchesTable + ` WHERE sent_at IS NULL AND letter_id=?)
	if q.stmtMarkLetterAsSent, err = conn.Prepare(`UPDATE ` + lettersTable + ` SET sent_at=?  WHERE id=? AND sent_at IS NULL AND NOT EXISTS (SELECT 1 FROM ` + messagesTable + ` WHERE letter_id=? AND sent_at IS NULL)`); err != nil {
		return nil, fmt.Errorf("unable to prepare mark letter as sent statement: %w", err)
	}
	if q.stmtDeleteLetter, err = conn.Prepare(`DELETE FROM ` + lettersTable + ` WHERE id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare delete letter statement: %w", err)
	}
	if q.stmtDeleteLetterAttachments, err = conn.Prepare(`DELETE FROM ` + attachmentsTable + ` WHERE letter_id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare delete letter attachments statement: %w", err)
	}
	if q.stmtDeleteLetterDispatches, err = conn.Prepare(`DELETE FROM ` + messagesTable + ` WHERE letter_id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare delete letter dispatches statement: %w", err)
	}

	if q.stmtListLettersForward, err = conn.Prepare(`SELECT id, frontmatter, content, created_at, sent_at FROM ` + lettersTable + ` WHERE id>? LIMIT ?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list letters statement: %w", err)
	}
	if q.stmtListLettersBackward, err = conn.Prepare(`SELECT id, frontmatter, content, created_at, sent_at FROM ` + lettersTable + ` WHERE id<? ORDER BY id DESC LIMIT ?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list letters statement: %w", err)
	}
	if q.stmtListAttachments, err = conn.Prepare(`SELECT name, content_hash, content_id, content_type, content FROM ` + attachmentsTable + ` WHERE letter_id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list attachments statement: %w", err)
	}
	if q.stmtListDispatchesForward, err = conn.Prepare(`SELECT id, headers, from_name, from_email, to_name, to_email, subject, message_text, message_html, sent_at FROM ` + messagesTable + ` WHERE letter_id=? AND id>? LIMIT ?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list dispatches statement: %w", err)
	}
	if q.stmtListDispatchesBackward, err = conn.Prepare(`SELECT id, headers, from_name, from_email, to_name, to_email, subject, message_text, message_html, sent_at FROM ` + messagesTable + ` WHERE letter_id=? AND id<? ORDER BY id DESC LIMIT ?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list dispatches statement: %w", err)
	}
	if q.stmtMarkMessageAsQueued, err = conn.Prepare(`UPDATE ` + messagesTable + ` SET queued_at=? WHERE id=? AND queued_at IS NULL`); err != nil {
		return nil, fmt.Errorf("unable to prepare mark message as queued statement: %w", err)
	}
	if q.stmtMarkMessageAsSent, err = conn.Prepare(`UPDATE ` + messagesTable + ` SET sent_at=? WHERE id=? AND sent_at IS NULL`); err != nil {
		return nil, fmt.Errorf("unable to prepare complete dispatch statement: %w", err)
	}

	return q, nil
}

func (q sqliteQueue) BindContext(ctx context.Context) func() {
	old := q.DB.SetInterrupt(ctx.Done())
	return func() {
		q.DB.SetInterrupt(old)
	}
}

func (q sqliteQueue) BeginTransaction(context.Context) (queue.Queue, queue.Transaction, error) {
	return q, transaction(sqlitex.Transaction(q.DB)), nil
}

func (q sqliteQueue) WithTransaction(
	ctx context.Context,
	tx queue.Transaction,
) (queue.Queue, error) {
	_, ok := tx.(transaction)
	if !ok {
		return nil, fmt.Errorf("imcompatible transaction type")
	}
	return q, nil
}

// escapeIdentifier safely quotes an SQLite table or column name.
func escapeIdentifier(name string) string {
	// Double quotes are escaped by doubling them in SQL identifiers
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

func encodeTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func decodeTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
