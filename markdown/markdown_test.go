package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFirstHeadingText(t *testing.T) {
	document, err := os.ReadFile(
		filepath.Join("testdata", "links.md"),
	)
	if err != nil {
		t.Fatal(err)
	}
	firstHeadingText := GetFirstHeadingText(document)
	if firstHeadingText != "Subtitle" {
		t.Fatal("unexpected first heading text:", firstHeadingText)
	}
}
