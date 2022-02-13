package main

import (
	"log"
	"mdsend"
	"os"

	"mdsend/distributors"
	"mdsend/gui"
	"mdsend/loaders"
	"mdsend/providers"
	"mdsend/renderers"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/spf13/cobra"
)

const appID = "com.github.dkotik.mdsend"

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
			o.Provider = providers.NewMailgunProvider(o.URI)

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

				application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
				if err != nil {
					log.Fatal("Could not create application:", err)
				}

				application.Connect("activate", func() {
					win, err := gui.Load()
					if err != nil {
						log.Fatal("Unable to create window:", err)
					}
					win.SetTitle("Markdown Send")
					win.SetLetter(string(message.Body))
					win.Connect("destroy", func() {
						application.Quit()
					})
					win.SetSensitive(true)
					win.SetApplication(application)
					win.SetDefaultSize(800, 500)

					for _, r := range *message.To {
						win.AddRecipient(r.String())
					}
					for _, r := range *message.CC {
						win.AddRecipient(r.String())
					}
					for _, r := range *message.BCC {
						win.AddRecipient(r.String())
					}

					win.Send.Connect(`clicked`, func() {
						win.Send.SetSensitive(false)
						// TODO: proper multi-tasking here
						go mdsend.Send(message, o)
					})
					win.ShowAll()
					o.Logger = win
				})
				// os.Args[0] = `MDSEND` // Title is generated from the exec. name
				// os.Exit(application.Run(os.Args))
				os.Exit(application.Run(nil))

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
