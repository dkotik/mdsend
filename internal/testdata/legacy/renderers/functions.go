package renderers

import (
	"encoding/base64"
	"html/template"
)

var templateFunctions = template.FuncMap{
	"base64":    base64.RawStdEncoding.EncodeToString,
	"base64URL": base64.RawURLEncoding.EncodeToString,
}
