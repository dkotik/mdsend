package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/urfave/cli/v3"
)

var flagQueue = &cli.StringFlag{
	Name:    `queue`,
	Usage:   `Path to the queue database file or data source name.`,
	Aliases: []string{`q`},
	Sources: cli.ValueSourceChain{
		Chain: []cli.ValueSource{
			cli.EnvVar("MDSEND_QUEUE"),
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

var application = &cli.Command{
	Name:      `mdsend`,
	Usage:     `Send markdown files as electronic mail. Maintain mailing lists as templated text files.`,
	Copyright: "Copyright 2022 Dmitry Kotik",
	Version:   version(),
	Flags: []cli.Flag{
		flagQueue,
		flagFrom,
		flagTo,
		flagGraceTimeout,
		flagDelay,
		flagFluctuate,
		flagWorkerCount,
		flagForever,
		flagDestroy,
		flagVerbose,
	},
	Commands: []*cli.Command{
		{
			Name:  `queue`,
			Usage: `Manages the queue of markdown documents to send.`,
			Commands: []*cli.Command{
				{
					Name:  `add`,
					Usage: `Adds a letter to the queue.`,
					Flags: []cli.Flag{
						// flagQueue,
						// flagFrom,
						// flagTo,
						// flagVerbose,
					},
					Action: cmdQueueAdd,
				},
				{
					Name:  `remove`,
					Usage: `Removes a letter from the queue.`,
					Flags: []cli.Flag{
						// flagQueue,
						// flagVerbose,
					},
					Action: func(ctx context.Context, c *cli.Command) error {
						return errors.New(`not implemented, yet`)
					},
				},
				{
					Name:  `view`,
					Usage: `Views the queue of markdown documents to send.`,
					Flags: []cli.Flag{
						// flagQueue,
					},
					Action: func(ctx context.Context, c *cli.Command) error {
						return errors.New(`not implemented, yet`)
					},
				},
				{
					Name:  `path`,
					Usage: `Prints the path to the current queue.`,
					Flags: []cli.Flag{
						// flagQueue,
					},
					Action: func(ctx context.Context, c *cli.Command) error {
						fmt.Println(c.String(flagQueue.Name))
						return nil
					},
				},
				{
					Name:  `clear`,
					Usage: `Removes all queued messages from the current queue.`,
					Flags: []cli.Flag{
						// flagQueue,
					},
					Action: func(ctx context.Context, c *cli.Command) error {
						p := c.String(flagQueue.Name)
						info, err := os.Stat(p)
						if err != nil {
							return err
						}
						if info.IsDir() {
							return errors.New("queue is a directory, not a file")
						}
						_ = os.Remove(p + ".db-wal")
						_ = os.Remove(p + ".db-shm")
						_ = os.Remove(p + ".db-journal")
						return os.Remove(p)
					},
				},
			},
		},
		{
			Name:  `validate`,
			Usage: `Validates the markdown document for correctness and readiness to be sent.`,
			Flags: []cli.Flag{
				// flagQueue,
				// flagVerbose,
			},
			Action: cmdValidate,
		},
		{
			Name:  `continue`,
			Usage: `Resume message delivery.`,
			Flags: []cli.Flag{
				// flagQueue,
				// flagVerbose,
			},
			Action: cmdSend,
		},
	},
	ExitErrHandler: func(
		ctx context.Context,
		cmd *cli.Command,
		err error,
	) {
		getLogger(cmd).Error(
			err.Error(),
		)
	},
	Action: cli.ActionFunc(func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() == 0 && !c.IsSet(flagForever.Name) {
			return cli.ShowRootCommandHelp(c) // nothing is happening
		}
		return cmdSend(ctx, c)
	}),
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	_ = application.Run(ctx, os.Args)
}

type xdgDataFile string

func (f xdgDataFile) Lookup() (string, bool) {
	v, err := xdg.DataFile("mdsend/" + string(f))
	if err != nil {
		return "", false
	}
	return v, true
}

func (f xdgDataFile) GoString() string {
	return "xdgDataFile{" + string(f) + "}"
}

func (f xdgDataFile) String() string {
	return string(f)
}

func version() string {
	v := "dev"
	if info, ok := debug.ReadBuildInfo(); ok {
		v = info.Main.Version
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				v = v + "-" + setting.Value
				break
			}
		}
	}
	return v
}
