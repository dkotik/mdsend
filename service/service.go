package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill-sqlite/wmsqlitezombiezen"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"golang.org/x/sync/errgroup"
	"zombiezen.com/go/sqlite"
)

type Options struct {
	Tracker   queue.ProgressTracker
	Retry     middleware.Retry
	Marshaler queue.Marshaler
	Logger    *slog.Logger
}

func New(
	ctx context.Context,
	group *errgroup.Group,
	router *message.Router,
	connectionDSN string,
	mailer mdsend.Mailer,
	options Options,
) (err error) {
	if group == nil {
		return errors.New("nil error group")
	}
	if options.Logger == nil {
		options.Logger = slog.Default()
	}
	if options.Tracker == nil {
		options.Tracker = queue.ProgressTrackerFunc(
			func(ctx context.Context, p queue.Progress) {
				options.Logger.Info(
					"messages sent",
					slog.String("count", fmt.Sprintf("%d/%d", p.Sent, p.Total)),
					slog.String("speed", fmt.Sprintf("%dm/s", p.MessagesPerSecond())),
					// slog.String("average", fmt.Sprintf("%.2fs", p.Average.Seconds())),
					slog.String("estimate_remaining", fmt.Sprintf("%.2fs", p.EstimateRemaining().Seconds())),
				)
			},
		)
	}
	if options.Marshaler == nil {
		options.Marshaler = queue.NewMarshalerJSON()
	}

	logger := options.Logger
	marshaler := options.Marshaler
	tracker := options.Tracker
	retry := options.Retry.Middleware
	group.Go(func() error {
		logger.Info("using database file", slog.String("path", connectionDSN))
		publisherConn, err := sqlite.OpenConn(connectionDSN)
		if err != nil {
			return fmt.Errorf("publisher database %q inaccessible: %w", connectionDSN, err)
		}
		defer func(conn *sqlite.Conn) {
			err = errors.Join(err, conn.Close())
		}(publisherConn)
		publisher, err := wmsqlitezombiezen.NewPublisher(publisherConn, wmsqlitezombiezen.PublisherOptions{
			InitializeSchema: true,
			Logger:           router.Logger(),
		})
		if err != nil {
			return fmt.Errorf("unable to setup database publisher: %w", err)
		}
		subscriber, err := wmsqlitezombiezen.NewSubscriber(connectionDSN, wmsqlitezombiezen.SubscriberOptions{
			PollInterval:     time.Millisecond * 30,
			InitializeSchema: true,
			Logger:           router.Logger(),
		})
		if err != nil {
			return fmt.Errorf("unable to setup database subscriber: %w", err)
		}

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
				retry(mailer),
				// message.HandlerFunc(func(msg *message.Message) ([]*message.Message, error) {
				// 	panic("djkflsjd")
				// }),
			)

			// a separate queue with its own connection is needed for each scheduler
			conn, err := sqlite.OpenConn(connectionDSN)
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

		queueConn, err := sqlite.OpenConn(connectionDSN)
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
			group,
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

		confirmationConn, err := sqlite.OpenConn(connectionDSN)
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
