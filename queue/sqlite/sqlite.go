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

	stmtInsertLetter   string
	stmtRetrieveLetter string
	stmtDeleteLetter   string
	stmtListLetters    string
}

// New creates an SQLite3 queue at the location.
func New(conn *sqlite.Conn, prefix string) (_ mdsend.Queue, err error) {
	if prefix == "" {
		prefix = "mdsend_"
	}
	lettersTable := escapeIdentifier(prefix + "letters")
	if err = sqlitex.ExecScript(
		conn,
		`
		CREATE TABLE IF NOT EXISTS `+lettersTable+` (
			id text PRIMARY KEY,
			frontmatter text,
			content text,
			created_at text,
			sent_at text
		);
		`,
	); err != nil {
		return nil, fmt.Errorf("unable to create tables: %w", err)
	}

	return queue{
		DB: conn,

		stmtInsertLetter:   `INSERT INTO ` + lettersTable + `(id, frontmatter, content, created_at, sent_at) VALUES(?,?,?,?,?)`,
		stmtRetrieveLetter: `SELECT frontmatter, content, created_at, sent_at FROM ` + lettersTable + ` WHERE id=?`,
		stmtDeleteLetter:   `DELETE FROM ` + lettersTable + ` WHERE id=?`,
		stmtListLetters:    `SELECT id, frontmatter, content, created_at, sent_at FROM ` + lettersTable,
	}, nil
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
