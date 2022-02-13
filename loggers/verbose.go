package loggers

import (
	"fmt"
	"log"
	"os"
)

type Verbose struct {
	sent, skipped, failed, total uint
}

func (v *Verbose) Open(s string) error {
	return nil
}

func (v *Verbose) Close() error {
	log.Print(v.Summary())
	return nil
}

func (v *Verbose) Summary() string {
	return fmt.Sprintf("%.2f%% delivered: %d sent, %d skipped, %d failed.", float32(v.sent)*100/float32(v.total), v.sent, v.skipped, v.failed)
}

func (v *Verbose) PeriodicSummary() {
	if (v.sent+v.skipped+v.failed)%20 == 0 { // Provide an update every 20 addresses.
		log.Print(".... " + v.Summary())
	}
}

// SetTotal adjusts the loggers total job count.
func (v *Verbose) SetTotal(jobs uint) {
	v.sent, v.skipped, v.failed, v.total = 0, 0, 0, jobs
}

func (v *Verbose) LogSent(s string, args ...interface{}) {
	v.sent++
	log.Printf("SENT "+s, args...)
	v.PeriodicSummary()
}

func (v *Verbose) LogSkip(s string, args ...interface{}) {
	v.skipped++
	log.Printf("SKIP "+s, args...)
	v.PeriodicSummary()
}

func (v *Verbose) LogInfo(s string, args ...interface{}) {
	log.Printf("INFO "+s, args...)
	v.PeriodicSummary()
}

func (v *Verbose) LogFail(s string, args ...interface{}) {
	v.failed++
	fmt.Fprintf(os.Stderr, "FAIL "+s+"\n", args...)
	v.PeriodicSummary()
}

func (v *Verbose) LogWarn(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "WARN "+s+"\n", args...)
}

func (v *Verbose) LogTest(s string, args ...interface{}) {
	v.skipped++
	log.Printf("TEST "+s, args...)
	v.PeriodicSummary()
}
