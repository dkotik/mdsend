package mailer

import (
	"sync"
	"testing"
	"time"

	"testing/synctest"

	"github.com/dkotik/mdsend"
)

func TestSemaphore(t *testing.T) {
	t.Parallel()

	delay := NewDelay(time.Second, time.Millisecond)
	s := NewSemaphore(
		delay(NewVoid()),
		delay(NewVoid()),
		delay(NewVoid()),
		delay(NewVoid()),
	)

	ctx := t.Context()
	begin := time.Now()
	wg := sync.WaitGroup{}
	for range 15 {
		wg.Add(1)
		go func() {
			_, _ = s.SendMail(ctx, mdsend.Message{})
			wg.Done()
		}()
	}
	wg.Wait()
	if time.Since(begin) < time.Second*3 {
		t.Errorf("expected duration to be at least 3 seconds, got %v", time.Since(begin))
	}

	synctest.Test(t, func(t *testing.T) {

	})
}
