package mdsend

import (
	"bytes"
	"os"
	"testing"

	"github.com/dkotik/mdsend/internal/media"
	"github.com/sebdah/goldie/v2"
	"gopkg.in/yaml.v3"
)

func TestLetterExtensions(t *testing.T) {
	letter, err := NewLetterFromFile(
		t.Context(),
		media.NewCyclicalImportPreventingFileSystem(
			os.DirFS("testdata/extend"),
		),
		"xedletter.md",
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
