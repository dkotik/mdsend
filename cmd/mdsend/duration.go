package main

import (
	"time"

	"github.com/dkotik/mdsend/internal/locale"
	"github.com/urfave/cli/v3"
)

var _ cli.ValueCreator[time.Duration, cli.NoConfig] = (*durationValue)(nil)

type extendedDurationFlag = cli.FlagBase[time.Duration, cli.NoConfig, durationValue]

// -- time.Duration Value
type durationValue time.Duration

func (d durationValue) Create(val time.Duration, p *time.Duration, c cli.NoConfig) cli.Value {
	*p = val
	return (*durationValue)(p)
}

func (d durationValue) ToString(val time.Duration) string {
	d = durationValue(val)
	return d.String()
}

func (d *durationValue) Set(s string) error {
	v, err := locale.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() any { return time.Duration(*d) }

func (d *durationValue) String() string {
	return locale.EncodeDuration(time.Duration(*d))
}

var (
	flagGraceTimeout = &extendedDurationFlag{
		Name:    "grace-timeout",
		Aliases: []string{`gt`},
		Value:   time.Second,
		Usage:   `Time allowance for the event router to finish current tasks when shutting down.`,
	}

	flagDelay = &extendedDurationFlag{
		Name:    `delay`,
		Aliases: []string{"d"},
		Usage:   `The minimum time delay between sending each electronic mail message.`,
	}

	flagFluctuate = &extendedDurationFlag{
		Name:    `fluctuate`,
		Aliases: []string{"f"},
		Usage:   `The time fluctuation in delay between sending each electronic mail message.`,
	}
)
