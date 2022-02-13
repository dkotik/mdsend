package providers

import (
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/go-multierror"
)

func NewMulticast(using ...Provider) Provider {
	queue := make(chan (*multicastJob), len(using)*2)

	return &multicastProvider{
		providers: using,
		queue:     queue,
		errors:    make(chan (error), len(using)*2),
		wg:        &sync.WaitGroup{},
	}
}

type multicastJob struct {
	DeliverTo string
	Message   io.ReadCloser
	Test      bool
}

type multicastProvider struct {
	providers []Provider
	queue     chan (*multicastJob)
	errors    chan (error)
	// lastError error
	wg *sync.WaitGroup
}

func (m *multicastProvider) Open() error {
	var result *multierror.Error
	for _, p := range m.providers {
		if err := p.Open(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if result != nil {
		return result
	}

	for _, p := range m.providers {
		go func(p Provider) {
			for job := range m.queue {
				// if job == nil {
				// 	m.wg.Done()
				// 	return // no more jobs coming
				// }
				resp, err := p.Send(job.DeliverTo, job.Message, job.Test)
				m.wg.Done()
				if err != nil {
					m.errors <- err
					return
				}
				fmt.Println("Multicast:", resp)
			}
		}(p)
	}
	return nil
}

func (m *multicastProvider) Send(to string, MIME io.ReadCloser, test bool) (string, error) {
	// requires waitgroup // TODO: here
	m.wg.Add(1)
	m.queue <- &multicastJob{
		DeliverTo: to,
		Message:   MIME,
		Test:      test,
	}
	return "queued one message to " + to, nil
}

func (m *multicastProvider) Close() error {
	m.wg.Wait() // wait for all sends to complete
	close(m.queue)
	var result *multierror.Error
	for err := range m.errors {
		result = multierror.Append(result, err)
	}
	close(m.errors)
	for _, p := range m.providers {
		if err := p.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}
