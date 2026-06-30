package main

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v3"
)

var flagVerbose = &cli.BoolFlag{
	Name:    "verbose",
	Aliases: []string{"v"},
	Usage:   "Enable verbose logging output.",
	Value:   false,
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
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: "2006-01-02 04:05",
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Value.Kind() == slog.KindAny {
					if _, ok := a.Value.Any().(error); ok {
						return tint.Attr(9, a) // make all errors red
					}
				}
				return a
			},
		}))
	}
	return logger
}
