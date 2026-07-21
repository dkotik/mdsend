package template

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strings"
)

// func (t *tmpl) Lookup(templateName string) bool {
// 	return t.HTML.Lookup(templateName) != nil
// }

func (t *tmpl) Execute(templateName string, data any) (v template.HTML, err error) {
	templateName = strings.TrimSpace(templateName)
	if templateName == "" {
		return "", errors.New("execute function requires a template name")
	}
	tmpl := t.HTML.Lookup(templateName)
	if tmpl == nil {
		return "", fmt.Errorf("no template %q found", templateName)
	}
	b := buffers.Get().(*bytes.Buffer)
	defer func(b *bytes.Buffer) {
		b.Reset()
		buffers.Put(b)
	}(b)
	if err = tmpl.Execute(b, data); err != nil {
		return "", fmt.Errorf("unable to execute template %q: %w", templateName, err)
	}
	return template.HTML(b.String()), nil
}

func (t *tmpl) Reify(templateName string) (v template.HTML, err error) {
	templateName = strings.TrimSpace(templateName)
	if templateName == "" {
		return "", errors.New("reify function requires a template name")
	}
	v, ok := t.ReifiedCache[templateName]
	if ok {
		return v, nil
	}
	v, err = t.Execute(templateName, t.context)
	if err != nil {
		return "", fmt.Errorf("unable to execute template %q: %w", templateName, err)
	}
	t.ReifiedCache[templateName] = v
	return v, nil
}

func (t *tmpl) Reset() {
	t.ReifiedCache = make(map[string]template.HTML)
}
