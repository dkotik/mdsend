package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/dkotik/mdsend"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type queue struct {
	DB *sqlite.Conn

	stmtInsertLetter            *sqlite.Stmt
	stmtInsertDispatch          *sqlite.Stmt
	stmtInsertAttachment        *sqlite.Stmt
	stmtRetrieveLetter          *sqlite.Stmt
	stmtDeleteLetter            *sqlite.Stmt
	stmtDeleteLetterAttachments *sqlite.Stmt
	stmtDeleteLetterDispatches  *sqlite.Stmt
	stmtListLettersForward      *sqlite.Stmt
	stmtListLettersBackward     *sqlite.Stmt
	stmtListAttachments         *sqlite.Stmt
	stmtListDispatchesForward   *sqlite.Stmt
	stmtListDispatchesBackward  *sqlite.Stmt
	stmtCompleteDispatch        *sqlite.Stmt
}

// New creates an SQLite3 queue at the location.
func New(conn *sqlite.Conn, prefix string) (_ mdsend.Queue, err error) {
	if prefix == "" {
		prefix = "mdsend_"
	}
	lettersTable := escapeIdentifier(prefix + "letters")
	attachmentsTable := escapeIdentifier(prefix + "attachments")
	dispatchesTable := escapeIdentifier(prefix + "dispatches")

	createLettersTable := true
	createAttachmentsTable := true
	createDispatchesTable := true

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
			case dispatchesTable:
				createDispatchesTable = false
			}
			return nil
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to list tables: %w", err)
	}

	if createLettersTable || createAttachmentsTable || createDispatchesTable {
		if err = sqlitex.ExecScript(
			conn,
			`
		CREATE TABLE IF NOT EXISTS `+lettersTable+` (
			id text PRIMARY KEY,
			frontmatter text NOT NULL,
			content text NOT NULL,
			created_at text NOT NULL,
			sent_at text NOT NULL
		) STRICT;

		CREATE INDEX IF NOT EXISTS index_name
			ON `+lettersTable+` (id);

		CREATE TABLE IF NOT EXISTS `+attachmentsTable+` (
			id text PRIMARY KEY,
			letter_id text NOT NULL,
			name text NOT NULL,
			content_hash text NOT NULL,
			content_type text NOT NULL,
			content blob NOT NULL,

			FOREIGN KEY (letter_id) REFERENCES `+lettersTable+`(id)
	   		ON DELETE CASCADE
	   		ON UPDATE CASCADE
		) STRICT;

		CREATE TABLE IF NOT EXISTS `+dispatchesTable+` (
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
			sent_at text
		) STRICT;
		`,
		); err != nil {
			return nil, fmt.Errorf("unable to create tables: %w", err)
		}
	}
	q := queue{
		DB: conn,
	}

	if q.stmtInsertLetter, err = conn.Prepare(`INSERT INTO ` + lettersTable + `(id, frontmatter, content, created_at, sent_at) VALUES(?,?,?,?,?)`); err != nil {
		return nil, fmt.Errorf("unable to prepare insert letter statement: %w", err)
	}
	if q.stmtInsertDispatch, err = conn.Prepare(`INSERT INTO ` + dispatchesTable + ` (id, letter_id, headers, from_name, from_email, to_name, to_email, subject, message_text, message_html) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`); err != nil {
		return nil, fmt.Errorf("unable to prepare insert dispatch statement: %w", err)
	}
	if q.stmtInsertAttachment, err = q.DB.Prepare(`INSERT INTO ` + attachmentsTable + ` (id, letter_id, name, content_hash, content_type, content) VALUES (?, ?, ?, ?, ?, ?)`); err != nil {
		return nil, fmt.Errorf("unable to prepare insert attachment statement: %w", err)
	}
	if q.stmtRetrieveLetter, err = conn.Prepare(`SELECT frontmatter, content, created_at, sent_at FROM ` + lettersTable + ` WHERE id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare retrieve letter statement: %w", err)
	}
	if q.stmtDeleteLetter, err = conn.Prepare(`DELETE FROM ` + lettersTable + ` WHERE id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare delete letter statement: %w", err)
	}
	if q.stmtDeleteLetterAttachments, err = conn.Prepare(`DELETE FROM ` + attachmentsTable + ` WHERE letter_id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare delete letter attachments statement: %w", err)
	}
	if q.stmtDeleteLetterDispatches, err = conn.Prepare(`DELETE FROM ` + dispatchesTable + ` WHERE letter_id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare delete letter dispatches statement: %w", err)
	}

	if q.stmtListLettersForward, err = conn.Prepare(`SELECT id, frontmatter, content, created_at, sent_at FROM ` + lettersTable + ` WHERE id>? LIMIT ?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list letters statement: %w", err)
	}
	if q.stmtListLettersBackward, err = conn.Prepare(`SELECT id, frontmatter, content, created_at, sent_at FROM ` + lettersTable + ` WHERE id<? ORDER BY id DESC LIMIT ?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list letters statement: %w", err)
	}
	if q.stmtListAttachments, err = conn.Prepare(`SELECT name, content_hash, content_type, content FROM ` + attachmentsTable + ` WHERE letter_id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list attachments statement: %w", err)
	}
	if q.stmtListDispatchesForward, err = conn.Prepare(`SELECT id, headers, from_name, from_email, to_name, to_email, subject, message_text, message_html, sent_at FROM ` + dispatchesTable + ` WHERE letter_id=? AND id>? LIMIT ?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list dispatches statement: %w", err)
	}
	if q.stmtListDispatchesBackward, err = conn.Prepare(`SELECT id, headers, from_name, from_email, to_name, to_email, subject, message_text, message_html, sent_at FROM ` + dispatchesTable + ` WHERE letter_id=? AND id<? ORDER BY id DESC LIMIT ?`); err != nil {
		return nil, fmt.Errorf("unable to prepare list dispatches statement: %w", err)
	}
	if q.stmtCompleteDispatch, err = conn.Prepare(`UPDATE ` + dispatchesTable + ` SET sent_at=? WHERE id=?`); err != nil {
		return nil, fmt.Errorf("unable to prepare complete dispatch statement: %w", err)
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
