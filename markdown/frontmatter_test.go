package markdown

import (
	"bytes"
	"testing"
)

func TestCut(t *testing.T) {
	tcs := []struct {
		Source      []byte
		Frontmatter []byte
		Content     []byte
		Delimeter   rune
	}{
		{
			Source: []byte(`
---
frontmatter
---
content
					`),
			Frontmatter: []byte(`frontmatter`),
			Content: []byte(`content
					`),
			Delimeter: '-',
		},
		{
			Source: []byte(`---
frontmatter

---
content

			`),
			Frontmatter: []byte(`frontmatter`),
			Content: []byte(`content

			`),
			Delimeter: '-',
		},
		{
			Source: []byte(`+++

frontmatter
+++
content
					`),
			Frontmatter: []byte(`frontmatter`),
			Content: []byte(`content
					`),
			Delimeter: '+',
		},
	}

	for i, tc := range tcs {
		frontmatter, content, delimeter, err := SplitFrontmatterFromContent(tc.Source)
		if err != nil {
			t.Errorf("%d. case failed: %v", i+1, err)
			continue
		}
		if delimeter != tc.Delimeter {
			t.Errorf("%d. delimeters did not match: %q vs %q", i+1, string(delimeter), string(tc.Delimeter))
			continue
		}
		if !bytes.Equal(frontmatter, tc.Frontmatter) {
			t.Errorf("%d. %s vs %s", i+1, frontmatter, tc.Frontmatter)
			continue
		}
		if !bytes.Equal(content, tc.Content) {
			t.Errorf("%d.  %s vs %s", i+1, content, tc.Content)
		}
	}
}
