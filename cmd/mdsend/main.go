package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"mdsend"
	"mdsend/distributors"
	"mdsend/loaders"
	"mdsend/loggers"
	"mdsend/providers"
	"mdsend/renderers"

	"github.com/spf13/cobra"
)

func main() {
	o := &mdsend.Options{
		Loader:      &loaders.ViperLoader{},
		Renderer:    &renderers.GoTemplateMIMERenderer{},
		Distributor: &distributors.LockingSynchronousBufferingDistributor{},
	}

	CLI := &cobra.Command{ // Setup command line interface.
		Use:     `mdsend`,
		Version: `0.2.1 Alpha`,
		Short:   `Sends markdown documents as templated emails.`,
		Long:    `Sends markdown documents as templated emails.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}

			// this needs to be fixed up a little bit!
			o.Provider = providers.NewMailgunProvider(o.URI)
			if o.Verbose {
				o.Logger = &loggers.Verbose{}
			} else {
				o.Logger = loggers.NewSpinner()
			}

			var answer string
			for _, mdfile := range args { // that's crazy should only do one at a time ==========================================
				handle, err := os.Open(mdfile)
				if err != nil {
					log.Fatalf(`Could not locate %s.`, mdfile)
				}
				message, err := o.Loader.Load(mdfile, handle)
				if err != nil {
					log.Fatalf(`Could not load %s, reason: %s.`, mdfile, err.Error())
				}
				handle.Close()
				if !o.YesOnPrompt {
					previewMessage(message)
					fmt.Print("Type \"yes\" (or \"test\") to confirm sending the message: ")
					fmt.Scanln(&answer)
					switch strings.ToLower(strings.TrimSpace(answer)) {
					case `test`:
						o.Dryrun = true
						fallthrough
					case `yes`:

					default:
						log.Fatal(`User cancelled delivery.`)
					}
				}
				mdsend.Send(message, o)
			}
		},
	}

	CLI.PersistentFlags().StringVarP(&o.URI, `uri`, `u`, os.Getenv(`MDSENDAPIURI`),
		`API URI for current provider. Defaults to environment variable MDSENDAPIURI.`)
	CLI.PersistentFlags().BoolVarP(&o.Verbose, `verbose`, `v`, false,
		`Print information about every message as it is being sent.`)
	CLI.PersistentFlags().BoolVarP(&o.YesOnPrompt, `yes`, `y`, false,
		`Automatically confirm all prompts.`)
	CLI.PersistentFlags().BoolVarP(&o.Dryrun, `dryrun`, `d`, false,
		`Print information about every message as it is being sent using API test mode without locking previously undelivered messages.`)
	CLI.Execute()
}
