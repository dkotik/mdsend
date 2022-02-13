package providers

import (
	"io"
)

// Provider delivers mail.
type Provider interface {
	Open() error
	Close() error
	Send(to string, MIME io.ReadCloser, test bool) (string, error)
}
