package template

import (
	"testing"

	"github.com/dkotik/mdsend"
)

func NewTemplateTest(tmpl Template) func(*testing.T) {
	return func(t *testing.T) {
		message, err := tmpl.RenderLetterForRecipient(nil)
		if err == nil {
			t.Fatal("expected failure for nil recipient")
		}
		message, err = tmpl.RenderLetterForRecipient(
			map[string]any{
				mdsend.FieldNameEmail: "testTo@test.com",
			},
		)
		if err == nil {
			t.Fatal("expected failure for empty recipient name")
		}
		message, err = tmpl.RenderLetterForRecipient(
			map[string]any{
				mdsend.FieldNameName: "testName",
			},
		)
		if err == nil {
			t.Fatal("expected failure for empty recipient address")
		}
		message, err = tmpl.RenderLetterForRecipient(
			map[string]any{
				mdsend.FieldNameName:  "testName",
				mdsend.FieldNameEmail: "testTo@test.com",
			},
		)
		if err != nil {
			t.Fatal("template rendering failed: ", err)
		}
		if err = message.Validate(); err != nil {
			// spew.Dump(tmpl)
			// spew.Dump(message.From)
			t.Fatal("template returned invalid message:", err)
		}
	}
}
