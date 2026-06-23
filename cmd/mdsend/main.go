package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	(&cli.Command{
		Name:    `mdsend`,
		Usage:   `Sends markdown documents as templated emails.`,
		Version: version(),
		Commands: []*cli.Command{
			{
				Name:    `queue`,
				Usage:   `Manages the queue of markdown documents to send.`,
				Aliases: []string{`q`},
				Commands: []*cli.Command{
					{
						Name:  `add`,
						Usage: `Adds a letter to the queue.`,
						Flags: []cli.Flag{
							flagDatabase,
							verboseFlag,
						},
						Action: cmdQueueAdd,
					},
					{
						Name:  `remove`,
						Usage: `Removes a letter from the queue.`,
						Flags: []cli.Flag{
							flagDatabase,
							verboseFlag,
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return errors.New(`not implemented, yet`)
						},
					},
					{
						Name:  `view`,
						Usage: `Views the queue of markdown documents to send.`,
						Flags: []cli.Flag{
							flagDatabase,
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return errors.New(`not implemented, yet`)
						},
					},
				},
			},
			{
				Name:    `send`,
				Usage:   `Sends markdown documents as templated emails.`,
				Aliases: []string{`s`},
				Flags: []cli.Flag{
					flagDatabase,
					flagGraceTimeout,
					verboseFlag,
					&cli.DurationFlag{
						Name:    `delay`,
						Aliases: []string{"d"},
						Usage:   `The minimum time delay between sending each electronic mail message.`,
					},
					&cli.DurationFlag{
						Name:    `fluctuate`,
						Aliases: []string{"f"},
						Usage:   `The time fluctuation in delay between sending each electronic mail message.`,
					},
					&cli.IntFlag{
						Name:    `worker_count`,
						Aliases: []string{"w"},
						Usage:   `The maximum number of simultaneous workers for sending electronic mail messages.`,
						Value:   max(1, runtime.NumCPU()),
						Action: func(ctx context.Context, c *cli.Command, v int) error {
							if v < 1 {
								return fmt.Errorf(`worker_count must be at least one: %d`, v)
							}
							if v > 64 {
								return fmt.Errorf(`worker_count exceeds 64: %d`, v)
							}
							return nil
						},
					},
				},
				Action: cmdSend,
			},
			{
				Name:    `validate`,
				Usage:   `Validates the markdown document for correctness and readiness to be sent.`,
				Aliases: []string{`t`},
				Flags: []cli.Flag{
					flagDatabase,
					verboseFlag,
				},
				Action: cmdValidate,
			},
		},
	}).Run(ctx, os.Args)
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
		v = `v` + info.Main.Version
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				v = v + "-" + setting.Value
				break
			}
		}
	}
	return v
}
