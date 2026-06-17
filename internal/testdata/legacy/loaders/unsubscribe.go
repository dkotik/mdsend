package loaders

import (
	"encoding/base64"
	"net/url"
	"text/template"

	"github.com/btcsuite/btcd/btcutil/base58"
)

func NewUnsubscribeLinkTemplate(URL string) (*template.Template, error) {
	return template.New("UnsubscribeLink").Funcs(template.FuncMap{
		"urlQuery": url.QueryEscape,
		"urlPath":  url.PathEscape,
		"base58": func(in string) string {
			return string(base58.Encode([]byte(in)))
		},
		"base64": func(in string) string {
			return base64.RawURLEncoding.EncodeToString([]byte(in))
		},
		"reverse": func(in string) string {
			n := 0
			rune := make([]rune, len(in))
			for _, r := range in {
				rune[n] = r
				n++
			}
			rune = rune[0:n]

			for i := 0; i < n/2; i++ { // reverse
				rune[i], rune[n-1-i] = rune[n-1-i], rune[i]
			}
			return string(rune)
		},
	}).Parse(URL)
}
