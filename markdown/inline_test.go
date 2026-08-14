package markdown

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestIndexImages(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "inline.md"))
	if err != nil {
		t.Fatal(err)
	}
	images := slices.Collect(IndexImages(NewParser(DefaultLightTheme), data))
	if len(images) != 3 {
		t.Fatalf("expected 3 images, got %d", len(images))
	}

	tcs := []string{"ref1URL", "img2URL", "ref3URL"}
	for i, tc := range tcs {
		if images[i].Location != tc {
			t.Errorf("expected %s, got %s", tc, images[i].Location)
		}
		cut := string(data[images[i].LocationAt : images[i].LocationAt+len(images[i].Location)])
		if tc != cut {
			t.Errorf("expected %s, got %s", tc, cut)
		}
	}
}
