package loader

import (
	"bytes"
	"os"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/sebdah/goldie/v2"
	"gopkg.in/yaml.v3"
)

func TestSplitOnLastHorizontalRule(t *testing.T) {
	data := []byte(`---
title: Test
---

* * *

_ _ _ _


tail`)
	most, tail, found := SplitOnLastHorizontalRule(data)
	if !found {
		t.Fatal("expected to find a horizontal rule")
	}
	if len(most) == 0 {
		t.Fatal("front of content was not found")
	}
	if len(tail) == 0 {
		t.Fatal("tail of content was not found")
	}
	if !bytes.Equal(tail, []byte(`tail`)) {
		t.Fatal("tail of content does not match expected")
	}
}

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
