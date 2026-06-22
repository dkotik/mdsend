package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func cmdSend(ctx context.Context, c *cli.Command) (err error) {
	if c.Args().Len() > 0 {
		if err = cmdQueueAdd(ctx, c); err != nil {
			return err
		}
	}
	return nil
}
