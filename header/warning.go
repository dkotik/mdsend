package header

import (
	"iter"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Warning struct {
	HeaderName string
	Message    *i18n.LocalizeConfig
}

func EachWarning(headers []Header) iter.Seq[Warning] {
	return func(yield func(Warning) bool) {
		known := make(map[string]int)
		for _, h := range headers {
			known[h.Name] = known[h.Name] + 1
		}

		for name, count := range known {
			if count < 2 {
				continue
			}
			if !yield(Warning{
				HeaderName: name,
				Message: &i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						Other: "there are {{.}} canonical header duplicates",
					},
					TemplateData: count - 1,
					PluralCount:  count - 1,
				},
			}) {
				return
			}
		}
	}
}
