package main

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

var flagVerbose = &cli.BoolFlag{
	Name:    "verbose",
	Aliases: []string{"v"},
	Usage:   "Enable verbose logging output.",
	Value:   true,
}

var logger *slog.Logger

func getLogger(ctx *cli.Command) *slog.Logger {
	if logger != nil {
		return logger
	}
	if ctx.Bool("verbose") {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.Level(slog.LevelDebug - 100),
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return logger
}
