package distributors

import (
	"crypto/rand"
	"log"

	"github.com/dkotik/mdsend/loaders"
	"github.com/dkotik/mdsend/loggers"
	"github.com/dkotik/mdsend/providers"
	"github.com/dkotik/mdsend/renderers"
)

// // ErrSkip is invoked by locking mechanisms, when the message had already been delivered to specified address.
// type ErrSkip struct {
// 	Hash uint64
// }
//
// func (e *ErrSkip) Error() string {
// 	return fmt.Sprintf(`letter with this subject and date had already been sent (lock #%x)`, e.Hash)
// }

// // ProgressFunction is called on every send.
// type ProgressFunction func(err error, email, message string)

// Distributor manages queues and workers to deliver mail.
type Distributor interface {
	// Progress(ProgressFunction)
	SetLogger(loggers.Logger)
	// When writing your own distributor, do not forget to set m.Current for each mail item!
	Send(l *loaders.Message, r renderers.Renderer, p providers.Provider, test bool) error
	Close() error
}

// RandomBytes provides N secure randomized bytes. Stops the program on failure.
func RandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil { // fires if len < n as well
		log.Fatal(`Unable to access secure random number generator.`)
	}
	return b
}
