package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sqlite/wmsqlitezombiezen"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/mailer"
	"github.com/dkotik/mdsend/mailer/environment"
	"github.com/dkotik/mdsend/queue"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
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

func cmdSend(ctx context.Context, c *cli.Command) (err error) {
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
	// mailer, err := newSemaphoreMailer(
	// 	c.Int(flagWorkerCount.Name),
	// 	middleware...,
	// )
	connectionDSN := c.String(flagQueue.Name)
	mailers := make([]mdsend.Mailer, c.Int(flagWorkerCount.Name))
	for i := range mailers {
		// m := mailer.NewVoid()
		conn, err := newQueueConnection(connectionDSN)
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
		for _, mw := range middleware {
			m = mw(m)
		}
		mailers[i] = m
	}
	// mailers = mailers[:1]
	mailer := mailer.NewSemaphore(mailers...)

	wg, ctx := errgroup.WithContext(ctx)
	// tracker := newProgressTracker(logger)
	tracker := newInterruptingProgressTracker(ctx, wg, logger)
	if err = send(
		ctx,
		wg,
		connectionDSN,
		c.Duration(flagGraceTimeout.Name),
		mailer,
		tracker,
		logger,
	); err != nil {
		return err
	}
	return wg.Wait()
}

func send(
	ctx context.Context,
	wg *errgroup.Group,
	connectionDSN string,
	graceTimeOut time.Duration,
	mailer mdsend.Mailer,
	tracker queue.ProgressTracker,
	logger *slog.Logger,
) (err error) {
	wg.Go(func() (err error) {
		// if err = addLetters(ctx, connectionDSN, []mdsend.Letter{
		// 	mdsend.Letter{
		// 		ID: "firstTestLetter",
		// 	},
		// }); err != nil {
		// 	return err
		// }
		logger.Info("using database file", slog.String("path", connectionDSN))
		wmLogger := watermill.NewSlogLoggerWithLevelMapping(
			logger,
			map[slog.Level]slog.Level{
				slog.LevelInfo: slog.LevelDebug,
			},
		)
		marshaler := queue.NewMarshalerJSON()

		publisherConn, err := newQueueConnection(connectionDSN)
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
			PollInterval:     time.Millisecond * 30,
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
				// message.HandlerFunc(func(msg *message.Message) ([]*message.Message, error) {
				// 	panic("djkflsjd")
				// }),
			)

			// a separate queue with its own connection is needed for each scheduler
			conn, err := newQueueConnection(connectionDSN)
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
			schedulers[i-1] = sqliteQ.NewScheduler(subQueue, marshaler, outbox, wmsqlitezombiezen.PublisherOptions{
				InitializeSchema: true,
			})
		}

		queueConn, err := newQueueConnection(connectionDSN)
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
		progressTracker := queue.NewProgressTracker(q, tracker)
		queue.NewContinuousScanner(
			ctx,
			wg,
			progressTracker,
			queue.NewRoundRobinScheduler(schedulers...),
			queue.ContinuousScannerOptions{
				// BeginWithOlderLetters: true,
				// Frequency:        time.Millisecond * 30,
				Frequency:        time.Second * 2,
				MessageBatchSize: 10,
			},
		)

		confirmationConn, err := newQueueConnection(connectionDSN)
		if err != nil {
			return fmt.Errorf("confirmation connection %q inaccessible: %w", connectionDSN, err)
		}
		defer func(conn *sqlite.Conn) {
			err = errors.Join(err, conn.Close())
		}(confirmationConn)
		confirmationQueue, err := sqliteQ.New(confirmationConn, "")
		if err != nil {
			return fmt.Errorf("unable to setup confirmation queue: %w", err)
		}
		router.AddConsumerHandler(
			"confirmation",
			"mdsendSent",
			subscriber,
			queue.NewConfirmationHandler(
				confirmationQueue,
				queue.ConfirmerFunc(func(ctx context.Context, c queue.Confirmation) error {
					// logger.Info("confirmation", slog.String("msg", c.MessageID), slog.String("confirmed", c.ID))
					return progressTracker.ConfirmScheduling(ctx, c)
				}), marshaler),
		)
		return router.Run(ctx)
	})

	return nil
}
