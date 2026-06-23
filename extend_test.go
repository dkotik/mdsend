package mdsend

import (
	"bytes"
	"os"
	"testing"

	"github.com/sebdah/goldie/v2"
	"gopkg.in/yaml.v3"
)

func TestLetterExtensions(t *testing.T) {
	letter, err := extend(
		t.Context(),
		Letter{
			Frontmatter: map[string]any{
				FieldNameSubject: "Test Letter Extensions",
				FieldNameExtends: []any{
					"first.yaml",
					"second.toml",
					"third.json",
					"fourth.cue",
					"fifth.md",
				},
			},
			Content: "Letter",
		},
		".",
		os.DirFS("testdata/extend"),
	)
	if err != nil {
		t.Fatal(err)
	}

	b := &bytes.Buffer{}
	_, _ = b.Write([]byte("---\n"))
	if err := yaml.NewEncoder(b).Encode(letter.Frontmatter); err != nil {
		t.Fatal(err)
	}
	_, _ = b.Write([]byte("---\n"))
	_, err = b.Write([]byte(letter.Content))
	if err != nil {
		t.Fatal(err)
	}

	goldie.New(t).Assert(t, "extend/extended", b.Bytes())
}
