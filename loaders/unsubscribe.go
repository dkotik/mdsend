package loaders

import (
	"encoding/base64"
	"net/url"
	"text/template"
)

func NewUnsubscribeLinkTemplate(URL string) (*template.Template, error) {
	return template.New("UnsubscribeLink").Funcs(template.FuncMap{
		"urlQuery": url.QueryEscape,
		"urlPath":  url.PathEscape,
		"base64": func(in string) string {
			return base64.RawURLEncoding.EncodeToString([]byte(in))
		},
	}).Parse(URL)
}
