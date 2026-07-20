package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"runtime"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/mailer"
	"github.com/dkotik/mdsend/mailer/environment"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"github.com/dkotik/mdsend/service"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
	"zombiezen.com/go/sqlite"
)

var (
	flagService = &cli.BoolFlag{
		Name:    `service`,
		Aliases: []string{"s"},
		Usage:   `Keep the process running and serving mail perpetually.`,
	}

	flagWorkerCount = &cli.IntFlag{
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
	}

	flagDestroy = &cli.BoolFlag{
		Name:   `destroy`,
		Usage:  `Destroy messages instead of sending them.`,
		Hidden: true,
	}
)

func cmdSend(ctx context.Context, c *cli.Command) (err error) {
	if c.Args().Len() > 0 {
		if err = cmdQueueAdd(ctx, c); err != nil {
			return err
		}
	}
	logger := getLogger(c)

	delay := c.Duration(flagDelay.Name)
	mailerMiddleware := []func(mdsend.Mailer) mdsend.Mailer{
		mailer.NewDelay(
			delay+time.Millisecond*50,
			c.Duration(flagFluctuate.Name)+time.Millisecond*20,
		),
	}
	if c.IsSet(flagFrom.Name) {
		mailerMiddleware = append(mailerMiddleware, mailer.NewFromOverride(
			c.Value(flagFrom.Name).(mail.Address),
		))
	}
	if c.Bool(flagVerbose.Name) {
		mailerMiddleware = append(mailerMiddleware, mailer.NewLogger(logger))
	}
	// mailer, err := newSemaphoreMailer(
	// 	c.Int(flagWorkerCount.Name),
	// 	mailerMiddleware...,
	// )
	connectionDSN := c.String(flagQueue.Name)
	mailers := make([]mdsend.Mailer, c.Int(flagWorkerCount.Name))
	destroy := c.Bool(flagDestroy.Name)
	for i := range mailers {
		conn, err := sqlite.OpenConn(connectionDSN)
		if err != nil {
			return fmt.Errorf("unable to connect to queue: %w", err)
		}
		defer conn.Close()
		queue, err := sqliteQ.New(conn, "")
		if err != nil {
			return fmt.Errorf("unable to connect to queue: %w", err)
		}

		m, err := environment.New(queue)
		if err != nil {
			return fmt.Errorf("unable to send mail: %w", err)
		}
		if destroy {
			m = mailer.NewVoid()
		}
		for _, mw := range mailerMiddleware {
			m = mw(m)
		}
		mailers[i] = m
	}

	wg, ctx := errgroup.WithContext(ctx)
	wmLogger := watermill.NewSlogLoggerWithLevelMapping(
		logger,
		map[slog.Level]slog.Level{
			slog.LevelInfo: slog.LevelDebug,
		},
	)
	router, err := message.NewRouter(
		message.RouterConfig{
			CloseTimeout: c.Duration(flagGraceTimeout.Name),
		},
		wmLogger,
	)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, router.Close())
	}()

	options := service.Options{
		Retry: middleware.Retry{
			InitialInterval: time.Second * 3,
			MaxInterval:     time.Minute * 10,
			// Multiplier is the factor by which the waiting interval will be multiplied between retries.
			Multiplier:          2.0,
			RandomizationFactor: 0.1,
			// OnRetryHook is an optional function that will be executed on each retry attempt.
			// The number of the current retry is passed as retryNum,
			// OnRetryHook func(retryNum int, delay time.Duration)
			ResetContextOnRetry: true,
			Logger:              wmLogger,
		},
	}
	if !c.Bool(flagService.Name) {
		// interrupt send command once everything appears to have
		// been sent
		options.Tracker = newInterruptingProgressTracker(
			ctx,
			wg,
			(time.Second*2)+delay,
			logger,
		)
	}

	if err = service.New(
		ctx,
		wg,
		router,
		connectionDSN,
		mailer.NewSemaphore(mailers...),
		options,
	); err != nil {
		return err
	}
	return wg.Wait()
}
