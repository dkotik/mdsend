package template

import (
	"bytes"
	"cmp"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/address"
	"github.com/dkotik/mdsend/header"
	"github.com/sebdah/goldie/v2"
	"golang.org/x/exp/slices"
)

func NewLetterTest(
	fs fs.FS,
	p string,
) func(*testing.T) {
	return func(t *testing.T) {
		letter, err := mdsend.NewLetterFromFile(t.Context(), fs, p)
		if err != nil {
			t.Fatal("unable to load letter:", err)
		}
		tmpl, err := New(letter, Options{})
		if err != nil {
			t.Fatal("unable to create template:", err)
		}
		for recipient, err := range address.Each(
			letter.Frontmatter,
			filepath.Dir(p),
			fs,
		) {
			if err != nil {
				t.Fatal("unable to load letter contacts:", err)
			}
			message, err := tmpl.RenderLetterForRecipient(recipient)
			if err != nil {
				t.Fatal("unable to execute template:", err)
			}
			b := &bytes.Buffer{}
			_, _ = io.WriteString(b, "subject: ")
			_, _ = io.WriteString(b, message.Subject)
			_ = b.WriteByte('\n')
			_, _ = io.WriteString(b, "seed_key: ")
			_, _ = io.WriteString(b, message.SeedKey)
			_ = b.WriteByte('\n')
			if !message.ScheduleAfter.IsZero() {
				_, _ = io.WriteString(b, "schedule_after: ")
				_, _ = io.WriteString(b, message.ScheduleAfter.Format(time.RFC3339))
				_ = b.WriteByte('\n')
			}
			slices.SortFunc(message.Headers, func(i, j header.Header) int {
				return cmp.Compare(i.Name, j.Name)
			})
			for _, header := range message.Headers {
				_, _ = io.WriteString(b, header.String())
				_ = b.WriteByte('\n')
			}
			_ = b.WriteByte('\n')
			_, _ = io.Copy(b, strings.NewReader(message.Text))
			_, _ = io.WriteString(b, "\n\n===========[HTML]=============\n\n")
			_, _ = io.Copy(b, strings.NewReader(message.HTML))

			ext := filepath.Ext(p)
			goldie.New(t, goldie.WithFixtureDir(filepath.Join("testdata", "examples"))).Assert(
				t,
				strings.TrimSuffix(filepath.Base(p), ext),
				b.Bytes(),
			)
			break
		}
	}
}

func NewTemplateTest(tmpl Template) func(*testing.T) {
	return func(t *testing.T) {
		message, err := tmpl.RenderLetterForRecipient(nil)
		if err == nil {
			t.Fatal("expected failure for nil recipient")
		}
		message, err = tmpl.RenderLetterForRecipient(
			map[string]any{
				address.FieldEmail: "testTo@test.com",
			},
		)
		if err == nil {
			t.Fatal("expected failure for empty recipient name")
		}
		message, err = tmpl.RenderLetterForRecipient(
			map[string]any{
				address.FieldName: "testName",
			},
		)
		if err == nil {
			t.Fatal("expected failure for empty recipient address")
		}
		message, err = tmpl.RenderLetterForRecipient(
			map[string]any{
				address.FieldName:  "testName",
				address.FieldEmail: "testTo@test.com",
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
