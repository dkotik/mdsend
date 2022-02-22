package mdsend

import (
	"log"
	"mdsend/distributors"
	"mdsend/loaders"
	"mdsend/loggers"
	"mdsend/providers"
	"mdsend/renderers"
)

const Version = "dev"

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

	// for i := 0; i < 1000; i++ {
	// 	switch rand.Intn(5) {
	// 	case 0:
	// 		logger.LogSkip(`skip %d here we go`, i)
	// 	case 1:
	// 		logger.LogWarn(`warn %d here we go`, i)
	// 	case 2:
	// 		logger.LogInfo(`info %d here we go`, i)
	// 	case 3:
	// 		logger.LogTest(`test %d here we go`, i)
	// 	default:
	// 		logger.LogSent(`sent %d here we go`, i)
	// 	}
	// 	time.Sleep(time.Second / 10)
	// }

	err := o.Distributor.Send(message, o.Renderer, o.Provider, o.Dryrun)
	if err != nil {
		o.Logger.LogFail(err.Error() + ".")
	}
	return err
}
