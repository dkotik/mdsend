package tests

import (
	"log"
	"os"
	"testing"

	"mdsend"
	"mdsend/distributors"
	"mdsend/loaders"
	"mdsend/providers"
	"mdsend/renderers"
)

func TestSendEmail(t *testing.T) {
	// log.Fatal(`env:`, os.Getenv(`MDSENDAPIURI`))
	o := mdsend.Options{
		URI:     os.Getenv(`MDSENDAPIURI`),
		Verbose: true,
		// Dryrun:      true,
		YesOnPrompt: true,
		Loader:      &loaders.ViperLoader{},
		Renderer:    &renderers.GoTemplateMIMERenderer{},
		Provider:    providers.NewMailgunProvider(os.Getenv(`MDSENDAPIURI`)),
		Distributor: &distributors.LockingSynchronousBufferingDistributor{},
	}
	err := mdsend.Send(`data/testemail.md`, &o)
	if err != nil {
		log.Fatal(err)
	}
	t.Fail()
}
