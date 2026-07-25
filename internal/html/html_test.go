package html

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/dkotik/mdsend/internal/locale"
	"golang.org/x/text/language"
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
			"isRTL": func(s string) (bool, error) {
				tag, err := language.Parse(s)
				if err != nil {
					return false, err
				}
				return locale.IsLanguageRightToLeft(tag), nil
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
