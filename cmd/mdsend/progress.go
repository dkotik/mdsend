package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/dkotik/mdsend/queue"
	"golang.org/x/sync/errgroup"
)

func newProgressTracker(
	logger *slog.Logger,
) queue.ProgressTracker {
	return queue.ProgressTrackerFunc(
		func(ctx context.Context, p queue.Progress) {
			logger.Info("progress", slog.Any("report", p))
		},
	)
}

func newInterruptingProgressTracker(
	ctx context.Context,
	eg *errgroup.Group,
	logger *slog.Logger,
) queue.ProgressTracker {
	if eg == nil {
		panic("nil error group")
	}
	if logger == nil {
		panic("nil logger")
	}
	closer := make(chan queue.Progress, 100)
	eg.Go(func() error {
		var p queue.Progress
		ticker := time.NewTicker(time.Second * 2)
		passed := 0
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case p = <-closer:
				passed = 0
			case <-ticker.C:
				if p.Sent < p.Total {
					continue
				}
				passed++
				if passed > 4 {
					logger.Info("progress tracker detected that everything had been sent, closing the program...")
					return context.Canceled
				}
			}
		}
	})
	return queue.ProgressTrackerFunc(
		func(ctx context.Context, p queue.Progress) {
			logger.Info("deliveries made", slog.Any("report", p))
			select {
			case <-ctx.Done():
				return
			case closer <- p:
			}
		},
	)
}
