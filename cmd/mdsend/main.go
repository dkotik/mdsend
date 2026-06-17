package main

import (
	"context"
	"os"
	"os/signal"
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
				Flags: []cli.Flag{
					flagDatabase,
				},
			},
			{
				Name:    `send`,
				Usage:   `Sends markdown documents as templated emails.`,
				Aliases: []string{`s`},
				Flags: []cli.Flag{
					flagDatabase,
				},
			},
			{
				Name:    `test`,
				Usage:   `Tests the markdown document for validity. Valid documents are ready to be sent.`,
				Aliases: []string{`t`},
				Flags: []cli.Flag{
					flagDatabase,
				},
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
