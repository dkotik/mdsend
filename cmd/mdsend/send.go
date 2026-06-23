package main

import (
	"context"
	"errors"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/urfave/cli/v3"
)

var flagGraceTimeout = &cli.DurationFlag{
	Name:    "grace-timeout",
	Aliases: []string{`gt`},
	Value:   time.Second,
}

func cmdSend(ctx context.Context, c *cli.Command) (err error) {
	if c.Args().Len() > 0 {
		if err = cmdQueueAdd(ctx, c); err != nil {
			return err
		}
	}
	logger := watermill.NewSlogLogger(getLogger(c))
	router, err := message.NewRouter(
		message.RouterConfig{
			CloseTimeout: c.Duration(flagGraceTimeout.Name),
		},
		logger,
	)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, router.Close())
	}()

	pubSub := gochannel.NewGoChannel(gochannel.Config{
		// Persistent: true,
	}, logger)
	// senderLogger := mailer.NewLogger(slogger)
	// MountSenders(
	// 	router,
	// 	pubSub,
	// 	pubSub,
	// 	"",
	// 	senderLogger(mailer.NewVoid()),
	// 	senderLogger(mailer.NewVoid()),
	// 	senderLogger(mailer.NewVoid()),
	// 	senderLogger(mailer.NewVoid()),
	// )

	router.AddConsumerHandler(
		"confirmation",
		"mdsend_confirmation",
		pubSub,
		func(msg *message.Message) error {
			// slogger.Info("confirmation", "msg", msg.UUID)
			// received++
			return nil
		},
	)

	// publisher := queue.NewRoundRobinPublisher(pubSub, "", 4)
	return router.Run(ctx)
}
