package markdown

import (
	"testing"
)

func TestAttachmentsDetection(t *testing.T) {
	result := CollectLinks([]byte(`![sm](perfect "sm")

# Title [](inTitle.jpg "In Title")

Inside paragraph: ![](inParagraph.jpg "In Paragraph") Tail of the paragraph.

In List:

- ![](inList.jpg "In List")


![flowing tail](finalMash1.txt "fm1")
[flowing tail](finalMash2.txt "fm2")

`))
	require := []Link{
		{Name: "sm", Path: "perfect"},
		{Name: "In Title", Path: "inTitle.jpg"},
		{Name: "In Paragraph", Path: "inParagraph.jpg"},
		{Name: "In List", Path: "inList.jpg"},
		{Name: "fm1", Path: "finalMash1.txt"},
		{Name: "fm2", Path: "finalMash2.txt"},
	}

	if len(result) != len(require) {
		t.Log("found:", result)
		t.Log("required:", require)
		t.Fatal("result does not match the required")
	}

	for i, link := range result {
		if link.Name != require[i].Name || link.Path != require[i].Path {
			t.Log("found:", result)
			t.Log("required:", require)
			t.Fatal("result does not match the required")
		}
	}
}
