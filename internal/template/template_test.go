package template

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/address"
	"github.com/dkotik/mdsend/internal/media"
	"github.com/sebdah/goldie/v2"
)

func TestValidMessageFromTemplate(t *testing.T) {
	tmpl, err := New(
		mdsend.Letter{
			ID: "valid message test",
			Frontmatter: map[string]any{
				address.FieldFrom: map[string]any{
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

func TestExamples(t *testing.T) {
	examplePath := filepath.Join("..", "..", "examples")
	entries, err := os.ReadDir(examplePath)
	if err != nil {
		t.Fatal("unable to reach examples folder:", err)
	}
	if len(entries) != 6 {
		t.Fatal("there should be 6 examples, instead:", len(entries))
	}
	fs := media.NewUnsafeUnconstrainedFileSystem()
	for _, entry := range entries {
		name := entry.Name()
		ext := filepath.Ext(name)
		if strings.ToLower(ext) != ".md" {
			continue // skip files that are not Markdown
		}
		t.Run(name, NewLetterTest(fs, filepath.Join(examplePath, name)))
	}
}
