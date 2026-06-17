package list

import (
	"context"
	"errors"

	"github.com/dkotik/mdsend/internal/locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func NewEmptyContactError(ctx context.Context) error {
	lc, ok := locale.LocalizerFromContext(ctx)
	if !ok {
		panic("no localizer")
	}
	return errors.New(lc.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "listContactEmptyError",
			Other: "cannot save empty contact",
		},
	}))
}

func NewEmptyNameError(ctx context.Context) error {
	lc, ok := locale.LocalizerFromContext(ctx)
	if !ok {
		panic("no localizer")
	}
	return errors.New(lc.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "listContactEmptyNameError",
			Other: "cannot use empty name",
		},
	}))
}

func NewEmptyEmailError(ctx context.Context) error {
	lc, ok := locale.LocalizerFromContext(ctx)
	if !ok {
		panic("no localizer")
	}
	return errors.New(lc.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "listContactEmptyEmailError",
			Other: "cannot use empty email address",
		},
	}))
}

func NewInvalidEmailError(ctx context.Context) error {
	lc, ok := locale.LocalizerFromContext(ctx)
	if !ok {
		panic("no localizer")
	}
	return errors.New(lc.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "listContactInvalidEmailError",
			Other: "invalid email address",
		},
	}))
}

func NewInvalidPhoneError(ctx context.Context) error {
	lc, ok := locale.LocalizerFromContext(ctx)
	if !ok {
		panic("no localizer")
	}
	return errors.New(lc.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "listContactInvalidPhoneError",
			Other: "invalid phone number",
		},
	}))
}

func NewUnsupportedSourceError(ctx context.Context, s string) error {
	lc, ok := locale.LocalizerFromContext(ctx)
	if !ok {
		panic("no localizer")
	}
	return errors.New(lc.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "listUnsupportedSourceTypeError",
			Other: "unsupported source type: {{ .Source }}",
		},
		TemplateData: map[string]interface{}{
			"Source": s,
		},
	}))
}
