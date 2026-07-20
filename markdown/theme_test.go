package markdown

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestThemeRendering(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "theme.md"))
	if err != nil {
		t.Fatal(err)
	}
	document := NewParser(DefaultLightTheme).Parse(text.NewReader(data))

	b := bytes.Buffer{}
	err = NewRendererHTML().Render(&b, data, document)
	if err != nil {
		t.Fatal(err)
	}

	goldie.New(t).Assert(t, "theme", b.Bytes())
}

func TestStyleInjection(t *testing.T) {
	document := NewParser(DefaultLightTheme).Parse(text.NewReader([]byte(`
# heading

par1

par2
		`)))

	err := ast.Walk(document, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		ApplyStyle(n, "color: #ff5555;")
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = ast.Walk(document, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		styleRaw, ok := n.Attribute(styleAttributeName)
		if !ok {
			return ast.WalkStop, fmt.Errorf("node missing style attribute: %s", n.Kind())
		}
		style := styleRaw.(string)
		if style != "color: #ff5555;" {
			return ast.WalkStop, fmt.Errorf("wrong node style attribute: %s", n.Kind())
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFontFamilyParsing(t *testing.T) {
	tcs := []string{
		`"Helvetica Neue", Arial, sans-serif`,
		`"Times New Roman", Times, serif`,
		`"Courier New", Courier, monospace`,
		`"Segoe UI", Tahoma, Geneva, Verdana, sans-serif`,
		`"Lucida Console", Monaco, monospace`,
		`"Gill Sans", "Gill Sans MT", Calibri, sans-serif`,
		`"Palatino Linotype", "Book Antiqua", Palatino, serif`,
		`"Trebuchet MS", Helvetica, sans-serif`,
		`"Brush Script MT", cursive`,
		`"Franklin Gothic Medium", "Arial Narrow", Arial, sans-serif`,
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			if !IsValidFontFamily(tc) {
				t.Fatalf("invalid font_family: %s", tc)
			}
		})
	}
}
