package template

import (
	"bytes"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/address"
)

func TestReifyTemplate(t *testing.T) {
	template, err := New(mdsend.Letter{
		Frontmatter: map[string]any{
			address.FieldFrom:       "Test <test@test.com>",
			mdsend.FieldNameSubject: "test",
		},
		Content: `
			{{- define "token" }}{{ RFC3339 }}{{ end -}}
			{{ reify "token" }}||{{ reify "token" -}}
		`,
	}, Options{})
	if err != nil {
		t.Fatal("cannot create template:", err)
	}

	synctest.Test(t, func(t *testing.T) {
		template := template.(*tmpl)
		template.Reset()
		b := &bytes.Buffer{}
		if err = template.Text.Execute(b, nil); err != nil {
			t.Fatal("unable to execute template")
		}
		early := b.String()
		first, second, ok := strings.Cut(early, "||")
		if !ok {
			t.Fatal("template did not render a pair of time tokens:", early)
		}
		if first != second {
			t.Fatal("first token does not match the second:", first, "vs", second)
		}
		<-time.After(time.Second)
		b.Reset()
		if err = template.Text.Execute(b, nil); err != nil {
			t.Fatal("unable to execute template")
		}
		later := b.String()
		if early != later {
			t.Fatal("reification failed")
		}
		first, second, ok = strings.Cut(later, "||")
		if !ok {
			t.Fatal("template did not render a pair of time tokens:", later)
		}
		if first != second {
			t.Fatal("first token does not match the second:", first, "vs", second)
		}
	})
}
