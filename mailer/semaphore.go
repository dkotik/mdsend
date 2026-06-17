package mailer

import (
	"context"

	"github.com/dkotik/mdsend"
)

type semaphore struct {
	available chan mdsend.Mailer
}

func NewSemaphore(available ...mdsend.Mailer) mdsend.Mailer {
	if len(available) == 0 {
		panic("at least one sender is required")
	}
	s := semaphore{
		available: make(chan mdsend.Mailer, len(available)),
	}
	for _, sender := range available {
		s.available <- sender
	}
	return s
}

func (s semaphore) SendMail(ctx context.Context, m mdsend.Message) (string, error) {
	sender := <-s.available
	defer func() {
		s.available <- sender
	}()
	return sender.SendMail(ctx, m)
}
