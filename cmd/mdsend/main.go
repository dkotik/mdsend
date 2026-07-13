package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/mailer"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

var application = &cli.Command{
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
						flagVerbose,
					},
					Action: cmdQueueAdd,
				},
				{
					Name:  `remove`,
					Usage: `Removes a letter from the queue.`,
					Flags: []cli.Flag{
						flagDatabase,
						flagVerbose,
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
				flagDelay,
				flagFluctuate,
				flagWorkerCount,
				flagVerbose,
			},
			Action: func(ctx context.Context, c *cli.Command) (err error) {
				if c.Args().Len() > 0 {
					if err = cmdQueueAdd(ctx, c); err != nil {
						return err
					}
				}
				logger := getLogger(c)

				middleware := []func(mdsend.Mailer) mdsend.Mailer{
					mailer.NewDelay(
						c.Duration(flagDelay.Name)+time.Millisecond*50,
						c.Duration(flagFluctuate.Name)+time.Millisecond*20,
					),
				}
				if c.Bool(flagVerbose.Name) {
					middleware = append(middleware, mailer.NewLogger(logger))
				}
				wg, ctx := errgroup.WithContext(ctx)
				if err = send(
					ctx,
					wg,
					c.String(flagDatabase.Name),
					c.Duration(flagGraceTimeout.Name),
					newSemaphoreMailer(
						c.Int(flagWorkerCount.Name),
						middleware...,
					),
					logger,
				); err != nil {
					return err
				}
				return wg.Wait()
			},
		},
		{
			Name:    `validate`,
			Usage:   `Validates the markdown document for correctness and readiness to be sent.`,
			Aliases: []string{`t`},
			Flags: []cli.Flag{
				flagDatabase,
				flagVerbose,
			},
			Action: cmdValidate,
		},
	},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	application.Run(ctx, os.Args)
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
