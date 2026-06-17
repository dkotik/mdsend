package sqlite

import (
	"context"

	"github.com/ThreeDotsLabs/watermill-sqlite/wmsqlitezombiezen"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/queue"
)

type scheduler struct {
	Queue     queue.Queue
	Publisher message.Publisher
	Topic     string
	Marshaler queue.Marshaler
}

func NewScheduler(q queue.Queue, m queue.Marshaler, topic string) queue.Scheduler {
	sq, ok := q.(sqliteQueue)
	if !ok {
		panic("queue is not an sqliteQueue")
	}
	pub, err := wmsqlitezombiezen.NewPublisher(sq.DB, wmsqlitezombiezen.PublisherOptions{})
	if err != nil {
		panic(err)
	}
	return scheduler{
		Queue:     q,
		Publisher: pub,
		Topic:     topic,
		Marshaler: m,
	}
}

func (s scheduler) ScheduleForDelivery(
	ctx context.Context,
	dispatch []mdsend.Dispatch,
) (err error) {
	if len(dispatch) == 0 {
		return nil
	}
	forPublisher := make([]*message.Message, 0, len(dispatch))
	ids := make([]string, 0, len(dispatch))
	for _, d := range dispatch {
		m, err := s.Marshaler.MarshalMessage(d)
		if err != nil {
			return err
		}
		// m.SetContext(ctx)
		forPublisher = append(forPublisher, m)
		ids = append(ids, d.ID)
	}
	forPublisher[0].SetContext(ctx)
	q, tx, err := s.Queue.BeginTransaction(ctx)
	defer tx.Close(&err)
	if err = q.MarkMessagesAsQueued(ctx, ids...); err != nil {
		return err
	}
	if err = s.Publisher.Publish(s.Topic, forPublisher...); err != nil {
		return err
	}
	return nil
}
