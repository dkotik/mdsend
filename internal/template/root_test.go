package template

import (
	"bytes"
	"html/template"
	"testing"
)

func TestReplaceRootTemplate(t *testing.T) {
	tmpl, err := template.New("").Parse(`
		[old root]
		{{ define "subTemplate" -}}
			[subTemplate]
		{{- end }}
		`)
	if err != nil {
		t.Fatalf("unable to create template: %v", err)
	}
	replaceWith, err := tmpl.New("").Parse(
		`[new root]`,
	)
	if err != nil {
		t.Fatal(err)
	}
	newTmpl := WithNewRootTemplate(tmpl, replaceWith)
	if newTmpl == nil {
		t.Fatalf("WithNewRootTemplate returned nil")
	}
	if newTmpl == tmpl {
		t.Fatalf("WithNewRootTemplate did not replace the root template")
	}

	b := &bytes.Buffer{}
	if err := newTmpl.Execute(b, nil); err != nil {
		t.Fatalf("unable to execute template: %v", err)
	}
	if b.String() != "[new root]" {
		t.Fatalf("WithNewRootTemplate did not replace the root template")
	}
}
