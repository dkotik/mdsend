package loader

import (
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestMessageLoading(t *testing.T) {
	m, err := NewMessage("../internal/testdata/pass/a.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(m.Path, "/internal/testdata/pass/a.md") {
		t.Fatalf("message path does not match: %q vs \"testdata/a.md\"", m.Path)
	}
	if m.Frontmatter == nil {
		t.Fatal("front-matter did not load")
	}
	if m.Content == "" {
		t.Fatal("content did not load")
	}
	if m.ID != "igJ3xax2yva" {
		t.Fatal("idempotent message ID does not match:", m.ID)
	}

	directory := filepath.Dir(m.Path) + "/"
	attachments := slices.Collect(maps.Values(m.Attachments))
	slices.Sort(attachments)
	for i, attachment := range attachments {
		attachments[i] = strings.TrimPrefix(attachment, directory)
	}

	if slices.Compare(
		attachments,
		[]string{
			"four.jpg",
			// "one.png",
			// "three.png",
			// "two.png",
		},
	) != 0 {
		t.Log(attachments)
		t.Fatal("attachments do not match the expected list")
	}
	// t.Fatal(m)
}
