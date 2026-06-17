package sender

import (
	"context"

	"github.com/dkotik/mdsend"
)

type semaphore struct {
	available chan mdsend.Sender
}

func NewSemaphore(available ...mdsend.Sender) mdsend.Sender {
	if len(available) == 0 {
		panic("at least one sender is required")
	}
	s := semaphore{
		available: make(chan mdsend.Sender, len(available)),
	}
	for _, sender := range available {
		s.available <- sender
	}
	return s
}

func (s semaphore) Send(ctx context.Context, m mdsend.Message) (string, error) {
	sender := <-s.available
	defer func() {
		s.available <- sender
	}()
	return sender.Send(ctx, m)
}
