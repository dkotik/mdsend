package template

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/address"
	"github.com/sebdah/goldie/v2"
)

func TestValidMessageFromTemplate(t *testing.T) {
	tmpl, err := New(
		mdsend.Letter{
			ID: "valid message test",
			Frontmatter: map[string]any{
				mdsend.FieldNameFrom: map[string]any{
					address.FieldName:  "Test Name From",
					address.FieldEmail: "from2from@test.com",
				},
				mdsend.FieldNameSubject: "test letter subject for {{ .Recipient.Name }}",
			},
			Content: "test letter for {{ .Recipient.Name }}",
		},
		Options{},
	)
	if err != nil {
		t.Fatal("unable to create template:", err)
	}
	if tmpl == nil {
		t.Fatal("nil template returned")
	}
	t.Run("interface conformity", NewTemplateTest(tmpl))
}

func TestTemplateRendering(t *testing.T) {
	defaultTemplate, err := defaultTemplates.ReadFile("html/default.html")
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("").Parse(string(defaultTemplate))
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	if err = tmpl.Execute(b, Context{
		Content: template.HTML("[test content]"),
	}); err != nil {
		t.Fatal(err)
	}

	goldie.New(t).Assert(t, "default", b.Bytes())
}
