package html

import (
	"bytes"
	"html/template"
	"testing"
)

func TestDefaultTemplateRendering(t *testing.T) {
	defaultTemplate := string(GetDefaultTemplateHTML())
	if defaultTemplate == "" {
		t.Fatal("default template is empty")
	}

	tmpl, err := template.New("").Funcs(
		map[string]any{
			"safeCSS": func(css string) template.CSS {
				return template.CSS(css)
			},
			"execute": func(templateName string, data any) template.HTML {
				return template.HTML("[execute:" + templateName + "]")
			},
		},
	).Parse(defaultTemplate)
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	if err = tmpl.Execute(b, nil); err != nil {
		t.Fatal(err)
	}

	// goldie.New(t).Assert(t, "default", b.Bytes())
}
