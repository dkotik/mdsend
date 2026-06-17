package queue

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/sender"
)

func TestQueue(t *testing.T) {
	slogger := slog.Default()
	logger := watermill.NewSlogLogger(slogger)
	router, err := message.NewRouter(
		message.RouterConfig{
			CloseTimeout: time.Second,
		},
		logger,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		router.Close()
	})

	// SignalsHandler will gracefully shutdown Router when SIGTERM is received.
	// You can also close the router by just calling `r.Close()`.
	router.AddPlugin(plugin.SignalsHandler)

	pubSub := gochannel.NewGoChannel(gochannel.Config{
		// Persistent: true,
	}, logger)
	senderLogger := sender.NewLogger(slogger)
	MountSenders(
		router,
		pubSub,
		pubSub,
		"",
		senderLogger(sender.NewVoid()),
		senderLogger(sender.NewVoid()),
		senderLogger(sender.NewVoid()),
		senderLogger(sender.NewVoid()),
	)

	received := 0
	router.AddConsumerHandler(
		"confirmation",
		"mdsend_confirmation",
		pubSub,
		func(msg *message.Message) error {
			slogger.Info("confirmation", "msg", msg.UUID)
			received++
			return nil
		},
	)

	publisher := NewRoundRobinPublisher(pubSub, "", 4)
	go func() {
		<-time.After(time.Second / 4)
		for range 10 {
			if err := publisher.Publish(context.Background(), mdsend.Message{
				ID: watermill.NewUUID(),
			}); err != nil {
				panic(err)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := router.Run(ctx); err != nil {
		panic(err)
	}
	if received != 10 {
		t.Fatalf("expected 10 received, got %d", received)
	}
}
