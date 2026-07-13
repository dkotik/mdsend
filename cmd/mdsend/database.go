package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
	"zombiezen.com/go/sqlite"
)

var flagDatabase = &cli.StringFlag{
	Name:    `database`,
	Usage:   `Path to the queue database file or data source name.`,
	Aliases: []string{`db`},
	Sources: cli.ValueSourceChain{
		Chain: []cli.ValueSource{
			cli.EnvVar("MDSEND_DATABASE"),
			xdgDataFile("queue.sqlite3"),
		},
	},
	Value: "mdsend_queue.sqlite3?cache=shared&foreign_keys=on",
	Validator: func(p string) error {
		if strings.TrimSpace(p) == "" {
			return errors.New(`database path is empty`)
		}
		return nil
	},
	Action: func(ctx context.Context, c *cli.Command, p string) (err error) {
		p = strings.TrimSpace(p)
		p, params, _ := strings.Cut(p, "?")
		dir := filepath.Dir(p)
		if _, err = os.Stat(dir); err != nil {
			// if os.IsNotExist(err) {
			// 	if err = os.MkdirAll(dir, 0700); err != nil {
			// 		return err
			// 	}
			// } else {
			// 	return err
			// }
			return fmt.Errorf("unable to access database directory %q: %w", dir, err)
		}
		paramValues := strings.Split(params, "&")
		if !slices.Contains(paramValues, `cache=shared`) {
			paramValues = append(paramValues, `cache=shared`)
		}
		if !slices.ContainsFunc(paramValues, func(v string) bool {
			return strings.HasPrefix(strings.TrimSpace(v), `foreign_keys=`)
		}) {
			paramValues = append(paramValues, `foreign_keys=on`)
		}
		c.Set(`database`, fmt.Sprintf("file:%s?%s", p, strings.Join(paramValues, "&")))
		// connectionDSN := "file:ephemeral?mode=memory&cache=shared"
		return nil
	},
}

func newDatabaseConnection(p string) (conn *sqlite.Conn, err error) {
	// userDir, err := os.UserHomeDir()
	// if err != nil {
	// 	return nil, err
	// }
	// p = "file:" + filepath.Join(userDir, "Downloads", "mdsend.sqlite3?cache=shared&wal=on")
	conn, err = sqlite.OpenConn(
		p,
		// flags ...sqlite.OpenFlags
	)
	if err != nil {
		return nil, err
	}
	// conn.SetBlockOnBusy()
	return conn, nil
}
