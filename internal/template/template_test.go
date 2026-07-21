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
	frontmatter := map[string]any{
		address.FieldFrom: map[string]any{
			address.FieldName:  "TestName",
			address.FieldEmail: "from2from@test.com",
		},
		mdsend.FieldNameSubject: "test letter subject for {{ .Recipient.Name }}",
	}
	tmpl, err := New(
		mdsend.Letter{
			ID:          "valid message test",
			Frontmatter: frontmatter,
			Content: `test letter for {{ .Recipient.name }}
				{{- if .IsPlainText -}}
					[plain text]
				{{- else -}}
					[html]
				{{- end -}}`,
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
	t.Run("IsPlainText value", func(t *testing.T) {
		message, err := tmpl.RenderLetterForRecipient(frontmatter[`from`].(map[string]any))
		if err != nil {
			t.Fatal("unable to render letter:", err)
		}
		if strings.Index(message.Text, "[plain text]") == -1 {
			t.Error("unexpected text content:", message.Text)
		}
		if strings.Index(message.HTML, "[html]") == -1 {
			// t.Fatal("unexpected html content:", message.HTML)
			t.Error("HTML content does not have the [html] tag")
		}
	})
}

func TestDefaultTemplateRendering(t *testing.T) {
	defaultTemplate, err := defaultTemplates.ReadFile("html/default.html")
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("").Funcs(functions()).Parse(string(defaultTemplate))
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
	if len(entries) != 7 {
		t.Fatal("there should be 7 examples, instead:", len(entries))
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
