package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/mailer"
	"github.com/dkotik/mdsend/queue"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"github.com/dkotik/watermillsqlite/wmsqlitezombiezen"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
	"zombiezen.com/go/sqlite"
)

var (
	flagGraceTimeout = &cli.DurationFlag{
		Name:    "grace-timeout",
		Aliases: []string{`gt`},
		Value:   time.Second,
	}

	flagDelay = &cli.DurationFlag{
		Name:    `delay`,
		Aliases: []string{"d"},
		Usage:   `The minimum time delay between sending each electronic mail message.`,
	}

	flagFluctuate = &cli.DurationFlag{
		Name:    `fluctuate`,
		Aliases: []string{"f"},
		Usage:   `The time fluctuation in delay between sending each electronic mail message.`,
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
)

func newSemaphoreMailer(
	size int,
	middleware ...func(mdsend.Mailer) mdsend.Mailer,
) mdsend.Mailer {
	mailers := make([]mdsend.Mailer, size)
	for i := range mailers {
		m := mailer.NewVoid()
		for _, mw := range middleware {
			m = mw(m)
		}
		mailers[i] = m
	}
	// mailers = mailers[:1]
	return mailer.NewSemaphore(mailers...)
}

func send(
	ctx context.Context,
	wg *errgroup.Group,
	connectionDSN string,
	graceTimeOut time.Duration,
	mailer mdsend.Mailer,
	logger *slog.Logger,
) (err error) {
	wg.Go(func() error {
		wmLogger := watermill.NewSlogLogger(logger)
		queueConn, err := sqlite.OpenConn(
			connectionDSN,
			// sqlite.OpenCreate, sqlite.OpenReadWrite,
		)
		if err != nil {
			return fmt.Errorf("queue database %q inaccessible: %w", connectionDSN, err)
		}
		defer func(conn *sqlite.Conn) {
			err = errors.Join(err, conn.Close())
		}(queueConn)
		q, err := sqliteQ.New(queueConn, "")
		if err != nil {
			return fmt.Errorf("unable to setup queue: %w", err)
		}

		publisherConn, err := sqlite.OpenConn(
			connectionDSN,
			// sqlite.OpenCreate, sqlite.OpenReadWrite,
		)
		if err != nil {
			return fmt.Errorf("publisher database %q inaccessible: %w", connectionDSN, err)
		}
		defer func(conn *sqlite.Conn) {
			err = errors.Join(err, conn.Close())
		}(publisherConn)
		publisher, err := wmsqlitezombiezen.NewPublisher(publisherConn, wmsqlitezombiezen.PublisherOptions{
			InitializeSchema: true,
			Logger:           wmLogger,
		})
		if err != nil {
			return fmt.Errorf("unable to setup database publisher: %w", err)
		}
		subscriber, err := wmsqlitezombiezen.NewSubscriber(connectionDSN, wmsqlitezombiezen.SubscriberOptions{
			InitializeSchema: true,
			Logger:           wmLogger,
		})
		if err != nil {
			return fmt.Errorf("unable to setup database subscriber: %w", err)
		}

		router, err := message.NewRouter(
			message.RouterConfig{
				CloseTimeout: graceTimeOut,
			},
			wmLogger,
		)
		if err != nil {
			return err
		}
		defer func() {
			err = errors.Join(err, router.Close())
		}()

		mailer := queue.NewSender(mailer)
		schedulers := make([]queue.Scheduler, 12)
		marshaler := queue.NewMarshalerJSON()
		for i := 1; i <= 12; i++ {
			outbox := fmt.Sprintf("mdsendOutbox%d", i)
			router.AddHandler(
				fmt.Sprintf("mdsendMailer%d", i),
				outbox,
				subscriber,
				"mdsendSent",
				publisher,
				// TODO: add retry
				mailer,
			)

			// a separate queue with its own connection is needed for each scheduler
			conn, err := sqlite.OpenConn(
				connectionDSN,
				// sqlite.OpenCreate, sqlite.OpenReadWrite,
			)
			if err != nil {
				return fmt.Errorf("outbox database %q inaccessible: %w", connectionDSN, err)
			}
			defer func(conn *sqlite.Conn) {
				err = errors.Join(err, conn.Close())
			}(conn)
			subQueue, err := sqliteQ.New(conn, "")
			if err != nil {
				return fmt.Errorf("unable to setup queue: %w", err)
			}
			schedulers[i-1] = sqliteQ.NewScheduler(subQueue, marshaler, outbox)
		}
		queue.NewContinuousScanner(ctx, wg, q, queue.NewRoundRobinScheduler(schedulers...), queue.ContinuousScannerOptions{
			Frequency: time.Millisecond * 30,
			// BeginWithOlderLetters: true,
		})
		// wg.Go(func() error {
		// 	for ids := range progress {
		// 		logger.Info("progress", slog.Any("ids", ids))
		// 	}
		// 	return nil
		// })

		confirmed := 0
		router.AddConsumerHandler(
			"confirmation",
			"mdsendSent",
			subscriber,
			func(msg *message.Message) error {
				confirmed++
				logger.Info("confirmation", slog.String("msg", msg.UUID), slog.Int("confirmed", confirmed))
				return nil
			},
		)
		return router.Run(ctx)
	})

	return nil
}
