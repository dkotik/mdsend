package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/mdsend"
)

type Publisher interface {
	Publish(context.Context, mdsend.Dispatch) error
}

type roundRobinPublisher struct {
	Publisher message.Publisher
	Buffer    *bytes.Buffer
	Topics    []string
	Current   int
}

func NewRoundRobinPublisher(publisher message.Publisher, prefix string, count uint) Publisher {
	if publisher == nil {
		panic("publisher is nil")
	}
	if count == 0 {
		panic("topic count is zero")
	}
	if prefix == "" {
		prefix = "mdsend"
	}
	topics := make([]string, count)
	for i := range topics {
		topics[i] = fmt.Sprintf("%s_outbox_%d", prefix, i+1)
	}
	return roundRobinPublisher{
		Publisher: publisher,
		Topics:    topics,
		Buffer:    &bytes.Buffer{},
		Current:   0,
	}
}

func (p roundRobinPublisher) Publish(ctx context.Context, dispatch mdsend.Dispatch) (err error) {
	p.Buffer.Reset()
	if err = json.NewEncoder(p.Buffer).Encode(dispatch); err != nil {
		return fmt.Errorf("failed to encode message to JSON: %w", err)
	}
	topic := p.Topics[p.Current]
	p.Current = (p.Current + 1) % len(p.Topics)

	m := message.NewMessage(dispatch.ID, p.Buffer.Bytes())
	m.SetContext(ctx)
	return p.Publisher.Publish(
		topic,
		m,
	)
}
