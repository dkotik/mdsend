package locale

import (
	"context"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func TestContextAwareness(t *testing.T) {
	bundle := i18n.NewBundle(language.English)
	localizer := i18n.NewLocalizer(bundle, "en")
	ctx := ContextWithLocalizer(context.Background(), localizer)
	if ctx == nil {
		t.Error("<nil> context")
	}

	recovered, ok := LocalizerFromContext(ctx)
	if !ok {
		t.Error("localizer was not recovered")
	}
	if recovered == nil {
		t.Error("recovered a <nil> localizer from context")
	}
}

func TestDefaultLangaugeTag(t *testing.T) {
	tag := language.Tag{}
	// tag := language.English
	if tag.String() != "und" {
		t.Fatal(tag.String())
	}

	if IsValidLanguageTag(tag) {
		t.Fatal("default language tag must be recognized as invalid")
	}
}
