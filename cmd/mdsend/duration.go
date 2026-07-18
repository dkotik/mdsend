package main

import (
	"time"

	"github.com/dkotik/mdsend/internal/locale"
	"github.com/urfave/cli/v3"
)

type extendedDurationFlag = cli.FlagBase[time.Duration, cli.NoConfig, durationValue]

// -- time.Duration Value
type durationValue time.Duration

// Below functions are to satisfy the ValueCreator interface

func (d durationValue) Create(val time.Duration, p *time.Duration, c cli.NoConfig) cli.Value {
	*p = val
	return (*durationValue)(p)
}

func (d durationValue) ToString(val time.Duration) string {
	d = durationValue(val)
	return d.String()
}

// Below functions are to satisfy the flag.Value interface

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
