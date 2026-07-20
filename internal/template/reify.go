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

func (t *tmpl) Reify(templateName string) (v template.HTML, err error) {
	templateName = strings.TrimSpace(templateName)
	if templateName == "" {
		return "", errors.New("reify function requires a template name")
	}
	v, ok := t.ReifiedCache[templateName]
	if ok {
		return v, nil
	}
	b := buffers.Get().(*bytes.Buffer)
	defer func(b *bytes.Buffer) {
		b.Reset()
		buffers.Put(b)
	}(b)
	tmpl := t.Text.Lookup(templateName)
	if tmpl == nil {
		return "", fmt.Errorf("no template %q found", templateName)
	}
	if err = tmpl.Execute(b, t.context); err != nil {
		return "", fmt.Errorf("unable to execute template %q: %w", templateName, err)
	}
	v = template.HTML(b.String())
	t.ReifiedCache[templateName] = v
	return v, nil
}

func (t *tmpl) Reset() {
	t.ReifiedCache = make(map[string]template.HTML)
}
