package main

import (
	"context"
	"log/slog"

	"github.com/dkotik/mdsend/queue"
	"golang.org/x/sync/errgroup"
)

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
	closer := make(chan queue.Progress, 1)
	eg.Go(func() error {
		var p queue.Progress
		passed := 0
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case p = <-closer:
				if p.Average == 0 && p.Sent == p.Total {
					passed++
					if passed > 2 {
						logger.Info("progress tracker detected that everything had been sent, closing the program...")
						return context.Canceled
					}
				} else {
					passed = 0
				}
			}
		}
	})
	return queue.ProgressTrackerFunc(
		func(ctx context.Context, p queue.Progress) {
			if p.Average > 0 { // at least some new deliveries were made
				logger.Info("deliveries made", slog.Any("report", p))
			}
			// logger.Warn("no new deliveries made", slog.Any("report", p))
			select {
			case <-ctx.Done():
				return
			case closer <- p:
			}
		},
	)
}
