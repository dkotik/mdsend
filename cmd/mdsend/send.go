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
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
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
	flagGraceTimeout = &extendedDurationFlag{
		Name:    "grace-timeout",
		Aliases: []string{`gt`},
		Value:   time.Second,
	}

	flagDelay = &extendedDurationFlag{
		Name:    `delay`,
		Aliases: []string{"d"},
		Usage:   `The minimum time delay between sending each electronic mail message.`,
	}

	flagFluctuate = &extendedDurationFlag{
		Name:    `fluctuate`,
		Aliases: []string{"f"},
		Usage:   `The time fluctuation in delay between sending each electronic mail message.`,
	}

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
)

func cmdSend(ctx context.Context, c *cli.Command) (err error) {
	if c.Args().Len() > 0 {
		if err = cmdQueueAdd(ctx, c); err != nil {
			return err
		}
	}
	logger := getLogger(c)

	delay := c.Duration(flagDelay.Name)
	middleware := []func(mdsend.Mailer) mdsend.Mailer{
		mailer.NewDelay(
			delay+time.Millisecond*50,
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
	tracker := newProgressTracker(logger)
	if !c.Bool(flagService.Name) {
		// interrupt send command once everything appears to have
		// been sent
		tracker = newInterruptingProgressTracker(
			ctx,
			wg,
			(time.Second*2)+delay,
			logger,
		)
	}

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
		retry := middleware.Retry{
			// MaxRetries is maximum number of times a retry will be attempted.
			// MaxRetries int

			// InitialInterval is the first interval between retries. Subsequent intervals will be scaled by Multiplier.
			InitialInterval: time.Second * 3,
			// MaxInterval sets the limit for the exponential backoff of retries. The interval will not be increased beyond MaxInterval.
			MaxInterval: time.Minute * 10,
			// Multiplier is the factor by which the waiting interval will be multiplied between retries.
			Multiplier: 2.0,
			// MaxElapsedTime sets the time limit of how long retries will be attempted. Disabled if 0.
			// MaxElapsedTime time.Duration
			// RandomizationFactor randomizes the spread of the backoff times within the interval of:
			// [currentInterval * (1 - RandomizationFactor), currentInterval * (1 + RandomizationFactor)].
			RandomizationFactor: 0.1,

			// OnRetryHook is an optional function that will be executed on each retry attempt.
			// The number of the current retry is passed as retryNum,
			// OnRetryHook func(retryNum int, delay time.Duration)

			// OnRetriesExhausted is an optional function that will be executed when all retries are exhausted.
			// This is called when MaxRetries is reached and the handler still returns an error.
			// It is NOT called when ShouldRetry returns false (that path returns a permanent error and exits earlier).
			// OnRetriesExhausted func(params RetriesExhaustedParams)

			// ResetContextOnRetry indicates whether the message context should be reset on each retry attempt.
			// See more: https://github.com/ThreeDotsLabs/watermill/issues/467
			//
			// This is not enabled by default to keep backward compatibility
			// (in theory, someone may want to preserve context values between retries).
			ResetContextOnRetry: true,
			Logger:              wmLogger,
		}.Middleware
		for i := 1; i <= 12; i++ {
			outbox := fmt.Sprintf("mdsendOutbox%d", i)
			router.AddHandler(
				fmt.Sprintf("mdsendMailer%d", i),
				outbox,
				subscriber,
				"mdsendSent",
				publisher,
				// TODO: add retry
				retry(mailer),
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
				Logger:           logger,
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
