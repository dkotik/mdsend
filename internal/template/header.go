package template

import (
	"fmt"
	"path"
	"text/template"

	"github.com/dkotik/mdsend"
)

type headerTemplate struct {
	Name     string
	Template *template.Template
}

func newBaseHeaderTemplate(
	l mdsend.Letter,
	functions template.FuncMap,
) (*template.Template, error) {
	base, err := template.New("content").Funcs(functions).Parse(l.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base template for headers: %w", err)
	}
	var subTemplate mdsend.Attachment
	for _, subTemplate = range l.Templates {
		base, err = base.New(
			path.Base(subTemplate.Name),
		).Parse(string(subTemplate.Content))
		if err != nil {
			return nil, fmt.Errorf("unable to parse template %q for headers: %w", subTemplate.Name, err)
		}
	}
	return base, nil
}

func newHeaderTemplateSet(
	l mdsend.Letter,
	baseTemplate *template.Template,
) (headers []headerTemplate, err error) {
	lheaders, err := l.GetHeaders()
	headers = make([]headerTemplate, len(lheaders))
	if err != nil {
		return nil, fmt.Errorf("unable to load headers: %w", err)
	}
	for i, header := range lheaders {
		clone, err := baseTemplate.Lookup("content").Clone()
		if err != nil {
			return nil, fmt.Errorf("unable to clone header template: %w", err)
		}
		value, err := clone.New("").Parse(header.Value)
		if err != nil {
			return nil, fmt.Errorf("unable to parse header %q as template: %w", header.Name, err)
		}
		headers[i] = headerTemplate{
			Name:     header.Name,
			Template: value,
		}
	}
	return headers, nil
}
