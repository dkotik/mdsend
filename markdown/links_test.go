package markdown

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestRelativePathPrefixInjection(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "links.md"))
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	if err = CopyWithRelativePathPrefix(b, source, "relativePathPrefix"); err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "prefixed", b.Bytes())
}

func TestAttachmentsDetection(t *testing.T) {
	source := `![sm](perfect "sm")

# Title [](inTitle.jpg "In Title")

Inside paragraph: ![](inParagraph.jpg "In Paragraph") Tail of the paragraph. And just a reference [refLink]!

In List:

- ![](inList.jpg "In List")


![flowing tail](finalMash1.txt "fm1")
[flowing tail](finalMash2.txt "fm2")

[refLink]:greatRefLink

`
	result := CollectLinks([]byte(source))
	require := []Link{
		{Name: "sm", Destination: "perfect"},
		{Name: "In Title", Destination: "inTitle.jpg"},
		{Name: "In Paragraph", Destination: "inParagraph.jpg"},
		{Name: "In List", Destination: "inList.jpg"},
		{Name: "fm1", Destination: "finalMash1.txt"},
		{Name: "fm2", Destination: "finalMash2.txt"},
		{Name: "", Destination: "greatRefLink"},
	}

	if len(result) != len(require) {
		t.Log("found:", result)
		t.Log("required:", require)
		t.Fatal("result does not match the required")
	}

	for i, link := range result {
		if link.Name != require[i].Name || link.Destination != require[i].Destination {
			t.Log("found:", require[i])
			t.Log("required:", link)
			t.Fatal("result does not match the required")
		}
		// if link.Destination != source[link.Position:link.Position+len(link.Destination)] {
		// 	t.Log("found:", source[link.Position:link.Position+len(link.Destination)])
		// 	t.Log("required:", link.Destination)
		// 	t.Fatal("result does not match the required")
		// }
	}
}

func TestReplaceDestinationForLinks(t *testing.T) {
	t.Skip("doodling")
	source := []byte(`[link](old "old")

[another][bottomRef]

[bottomRef]: anotherRef

`)
	pc := parser.NewContext()
	ast.Walk(
		attachmentsSniffer.Parse(
			text.NewReader(source),
			parser.WithContext(pc),
		),
		ast.Walker(func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				switch n := n.(type) {
				case *ast.LinkReferenceDefinition:
					t.Logf("found link def: %s %v", n.Destination, n.Pos())
				case *ast.Link:
					// n.
					// t.Logf("found link: %s %d", n.Destination, n.ChildCount())
					t.Logf("link pos: %v", n.Pos())
					// 	n.Attributes()
					// 	lines := n.Lines()
					// 	for i := 0; i < lines.Len(); i++ {
					// 		line := lines.At(i)
					// 		t.Log(string(source[line.Start:line.Stop]))
					// 	}
					// 	// lines := n.Lines().At(i int)
					// 	// result = append(result, Link{Name: string(n.Title), Destination: string(n.Destination)})
					// case *ast.Image:
					// 	// result = append(result, Link{Name: string(n.Title), Destination: string(n.Destination)})
				}
			}
			return ast.WalkContinue, nil
		}),
	)
	// for _, ref := range pc.References() {
	// 	t.Logf("ref: %s %s", ref.Label(), ref.Destination())
	// }
	t.Fatal("expected to find link reference definition")
}
