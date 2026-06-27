package queue

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/mdsend"
	"github.com/google/uuid"
)

type Marshaler interface {
	MarshalMessage(any) (*message.Message, error)
}

type marshalerJSON struct{}

func NewMarshalerJSON() Marshaler {
	return marshalerJSON{}
}

func (m marshalerJSON) MarshalMessage(v any) (*message.Message, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return message.NewMessage(uuid.NewString(), payload), nil
}

type Scheduler interface {
	ScheduleForDelivery(context.Context, mdsend.Letter, []mdsend.Message) error
}

type SchedulerFunc func(context.Context, mdsend.Letter, []mdsend.Message) error

func (f SchedulerFunc) ScheduleForDelivery(
	ctx context.Context,
	letter mdsend.Letter,
	batch []mdsend.Message,
) error {
	return f(ctx, letter, batch)
}

type basicScheduler struct {
	Queue     Queue
	Publisher message.Publisher
	Topic     string
	Marshaler Marshaler
}

// NewSchedulerForPublisher returns a basic transaction-less scheduler. It marks the messsage as queued first
// to avoid duplicate deliveries. If the publisher fails to publish, the message is dropped. The sacrifice
// of consistency allows the scheduler to use the queue and publisher that do not use a compatible database
// driver.
//
// Queue drivers sometimes provide a stricter scheduler.
func NewSchedulerForPublisher(q Queue, m Marshaler, pub message.Publisher, topic string) Scheduler {
	if q == nil {
		panic("queue is nil")
	}
	if pub == nil {
		panic("publisher is nil")
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		panic("topic is empty")
	}
	if m == nil {
		m = NewMarshalerJSON()
	}
	return basicScheduler{
		Queue:     q,
		Publisher: pub,
		Topic:     topic,
		Marshaler: m,
	}
}

func (s basicScheduler) ScheduleForDelivery(
	ctx context.Context,
	l mdsend.Letter,
	m []mdsend.Message,
) (err error) {
	if len(m) == 0 {
		return nil
	}
	forPublisher := make([]*message.Message, 0, len(m))
	ids := make([]string, 0, len(m))
	for _, d := range m {
		m, err := s.Marshaler.MarshalMessage(d)
		if err != nil {
			return err
		}
		// m.SetContext(ctx)
		forPublisher = append(forPublisher, m)
		ids = append(ids, d.ID)
	}
	forPublisher[0].SetContext(ctx)
	if err = s.Queue.MarkMessagesAsScheduled(ctx, l.ID, ids...); err != nil {
		return err
	}
	if err = s.Publisher.Publish(s.Topic, forPublisher...); err != nil {
		return err
	}
	return nil
}

type roundRobinScheduler struct {
	Cursor     int
	Schedulers []Scheduler
}

// NewRoundRobinScheduler rotates through the provided schedulers on each scheduled messsage.
func NewRoundRobinScheduler(schedulers ...Scheduler) Scheduler {
	if len(schedulers) == 0 {
		panic("no schedulers were provided")
	}
	for _, s := range schedulers {
		if s == nil {
			panic("nil scheduler provided")
		}
	}
	return &roundRobinScheduler{
		Cursor:     -1,
		Schedulers: schedulers,
	}
}

func (s *roundRobinScheduler) ScheduleForDelivery(
	ctx context.Context,
	l mdsend.Letter,
	m []mdsend.Message,
) error {
	s.Cursor = (s.Cursor + 1) % len(s.Schedulers)
	return s.Schedulers[s.Cursor].ScheduleForDelivery(ctx, l, m)
}
