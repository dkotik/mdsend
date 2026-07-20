package markdown

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestActionParsingAndRendering(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "actions.md"))
	if err != nil {
		t.Fatal(err)
	}
	parser := NewParser(DefaultLightTheme)
	document := parser.Parse(text.NewReader(data))

	expectActions := 3
	foundActions := 0
	if err = ast.Walk(document, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering == true && n.Kind() == KindActionButton {
			n := n.(*ActionButtonNode)
			t.Log("action:", string(n.Link.Destination), string(n.Link.Title))
			foundActions++
		}
		return ast.WalkContinue, nil
	}); err != nil {
		t.Fatal(err)
	}

	if expectActions != foundActions {
		t.Fatal("action count mismatch:", expectActions, "vs", foundActions)
	}

	renderer := NewRendererHTML()

	b := &bytes.Buffer{}
	if err = renderer.Render(b, data, document); err != nil {
		t.Fatal(err)
	}

	goldie.New(t).Assert(t, "actions.html", b.Bytes())
}
