package template

import "text/template"

type headerTemplate struct {
	Name     string
	Template *template.Template
}
