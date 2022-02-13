package loggers

import (
	"fmt"

	"github.com/leaanthony/spinner"
)

var globalSpinner *spinner.Spinner

func resetSpinner() {
	globalSpinner = spinner.New()
	globalSpinner.SetSpinSpeed(60)
	globalSpinner.SetAbortMessage("Operation cancelled.")
	globalSpinner.SetSpinFrames([]string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"})
	globalSpinner.Start()
}

func NewSpinner() *Spinner {
	return &Spinner{Verbose{}}
}

// Spinner displays current delivery status.
type Spinner struct {
	Verbose
}

func (sp *Spinner) Open(s string) error {
	resetSpinner()
	return nil
}

func (sp *Spinner) Close() error {
	globalSpinner.Success("Done: " + sp.Verbose.Summary())
	return nil
}

func (sp *Spinner) update(s string, args ...interface{}) {
	// time.Sleep(time.Second)
	globalSpinner.UpdateMessage(sp.Verbose.Summary() + " " + fmt.Sprintf(s, args...))
}

func (sp *Spinner) freeze(s string, args ...interface{}) {
	globalSpinner.Errorf(s, args...)
	resetSpinner()
	sp.update(s, args...)
}

func (sp *Spinner) LogSent(s string, args ...interface{}) {
	sp.Verbose.sent++
	// sp.update("SENT "+s, args...)
	if len(args) > 0 {
		sp.update("Sent to %s.", args[0])
	} else {
		sp.update("Sent.")
	}
}

func (sp *Spinner) LogSkip(s string, args ...interface{}) {
	sp.Verbose.skipped++
	sp.freeze("Skipped: "+s, args...)
}

func (sp *Spinner) LogTest(s string, args ...interface{}) {
	sp.update("TEST "+s, args...)
}

func (sp *Spinner) LogFail(s string, args ...interface{}) {
	sp.Verbose.failed++
	sp.freeze("Error: "+s, args...)
}

func (sp *Spinner) LogInfo(s string, args ...interface{}) {
	sp.update(s, args...)
}

func (sp *Spinner) LogWarn(s string, args ...interface{}) {
	sp.update("Warning: "+s, args...)
}
