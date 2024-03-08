package locale

import (
	"context"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

// LocalizerFromContext raises request-scoped localizer from context.
// Returns `<nil>, false` if there is no localizer in context.
func LocalizerFromContext(ctx context.Context) (l *i18n.Localizer, ok bool) {
	l, ok = ctx.Value(contextKey).(*i18n.Localizer)
	return
}

// ContextWithLocalizer adds localizer into context as a value.
// Use [LocalizerFromContext] to recover it later.
func ContextWithLocalizer(parent context.Context, l *i18n.Localizer) context.Context {
	return context.WithValue(parent, contextKey, l)
}
