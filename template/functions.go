package template

import (
	"encoding/base64"
	"html/template"
)

func Functions() template.FuncMap {
	return template.FuncMap{
		"base64":    base64.RawStdEncoding.EncodeToString,
		"base64URL": base64.RawURLEncoding.EncodeToString,
	}
}
