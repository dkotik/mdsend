package locale

import (
	"errors"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Localizable interface {
	Localize(*i18n.Localizer) (string, error)
}

func LocalizeError(err error, lc *i18n.Localizer) string {
	localizable, ok := err.(Localizable)
	if ok {
		message, lerr := localizable.Localize(lc)
		if lerr != nil {
			return errors.Join(err, lerr).Error()
		}
		return message
	}
	return err.Error()
}
