package mdsend

import (
	"context"
	"log"

	"github.com/dkotik/mdsend/distributors"
	"github.com/dkotik/mdsend/loaders"
	"github.com/dkotik/mdsend/loggers"
	"github.com/dkotik/mdsend/providers"
	"github.com/dkotik/mdsend/renderers"
)

const Version = "dev"

type Sender interface {
	Send(context.Context, Dispatch) (string, error)
}

// Options provide configuration to component execution.
type Options struct {
	URI         string // Credentials for provider API backend.
	Verbose     bool
	Dryrun      bool // Run the program without delivering or locking anything down.
	YesOnPrompt bool // Automatically confirm all prompts.
	Logger      loggers.Logger
	Loader      loaders.Loader
	Renderer    renderers.Renderer
	Provider    providers.Provider
	Distributor distributors.Distributor
}

// Send a message with default agents.
func Send(message *loaders.Message, o *Options) error {
	if message.Subject == `` {
		log.Fatal(`Message subject must be specified in the markdown file!`)
	}
	o.Logger.SetTotal(uint(len(*message.To) + len(*message.CC) + len(*message.BCC)))
	o.Distributor.SetLogger(o.Logger)
	o.Logger.Open("")
	defer o.Logger.Close()
	defer o.Distributor.Close()

	err := o.Distributor.Send(message, o.Renderer, o.Provider, o.Dryrun)
	if err != nil {
		o.Logger.LogFail(err.Error() + ".")
	}
	return err
}
