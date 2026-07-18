package markdown

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/yuin/goldmark"
)

func TestPlaintextRenderer(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "links.md"))
	if err != nil {
		t.Fatal(err)
	}
	md := goldmark.New(
		goldmark.WithRenderer(NewPlaintextRenderer()),
		goldmark.WithParser(NewParser()),
	)

	b := &bytes.Buffer{}
	if err := md.Convert(source, b); err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "plaintext", b.Bytes())
}
