package loader

import (
	"bytes"
	"os"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/sebdah/goldie/v2"
	"gopkg.in/yaml.v3"
)

func TestLetterExtensions(t *testing.T) {
	loader := loader{
		FS:    os.DirFS("testdata/extend"),
		Cache: NewMapCache(),
	}
	letter, err := loader.extend(
		t.Context(),
		mdsend.Letter{
			Frontmatter: map[string]any{
				mdsend.FieldNameSubject: "Test Letter Extensions",
				mdsend.FieldNameExtends: []any{
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
		make(map[string]struct{}),
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
