package main

import (
	"net/mail"

	"github.com/dkotik/mdsend/address"
	"github.com/urfave/cli/v3"
)

var _ cli.ValueCreator[mail.Address, cli.NoConfig] = (*addressValue)(nil)

type (
	addressValue     mail.Address
	addressFlag      = cli.FlagBase[mail.Address, cli.NoConfig, addressValue]
	addressSlice     = cli.SliceBase[mail.Address, cli.NoConfig, addressValue]
	addressSliceFlag = cli.FlagBase[[]mail.Address, cli.NoConfig, addressSlice]
)

func (d addressValue) Create(val mail.Address, p *mail.Address, c cli.NoConfig) cli.Value {
	*p = val
	return (*addressValue)(p)
}

func (d addressValue) ToString(val mail.Address) string {
	return d.String()
}

func (d *addressValue) Set(s string) error {
	v, err := mail.ParseAddress(s)
	if err != nil {
		return err
	}
	if err = address.ValidateFormat(v.Address); err != nil {
		return err
	}
	*d = addressValue(*v)
	return err
}

func (d *addressValue) Get() any { return mail.Address(*d) }

func (d *addressValue) String() string {
	return d.String()
}

var (
	flagFrom = &addressFlag{
		Name:    "from",
		Aliases: []string{`f`},
		Usage:   `Override message author when queuing or sending messages.`,
	}

	flagTo = &addressSliceFlag{
		Name:    "to",
		Aliases: []string{`cc`, `bcc`},
		Usage:   `Add additional recipients when queuing or sending messages.`,
	}
)
