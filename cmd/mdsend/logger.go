package main

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

var verboseFlag = &cli.BoolFlag{
	Name:    "verbose",
	Aliases: []string{"v"},
	Usage:   "Enable verbose logging output.",
	Value:   false,
}

func getLogger(ctx *cli.Command) *slog.Logger {
	if ctx.Bool("verbose") {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.Level(slog.LevelDebug - 100),
		}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}
