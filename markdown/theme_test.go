package markdown

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func TestThemeRendering(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "theme.md"))
	if err != nil {
		t.Fatal(err)
	}
	document := parser.NewParser(parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(
			parser.DefaultParagraphTransformers()...,
		),
		parser.WithASTTransformers(
			// util.Prioritized(&ActionButtonInjector{}, 100),
			util.Prioritized(Theme{
				Color: Color{
					Action:     "#5454",
					Heading:    "#5454",
					Text:       "#5454",
					Link:       "#5454",
					BlockQuote: "#5454",
					Border:     "#5454",
					Table:      "#5454",
					Shadow:     "#5454",
				},
				FontFamily: "Georgia",
				FontSize:   12,
			}, 1000),
		),
	).Parse(text.NewReader(data))

	b := bytes.Buffer{}
	err = renderer.NewRenderer(
		renderer.WithNodeRenderers(
			// util.Prioritized(&htmlRenderer{
			// 	BlockQuoteStyle:      options.BlockQuoteStyle,
			// 	LinkStyle:            options.LinkStyle,
			// 	ActionContainerStyle: options.ActionContainerStyle,
			// 	ActionStyle:          options.ActionStyle,
			// 	Writer:               html.DefaultWriter,
			// }, 1000),
			util.Prioritized(html.NewRenderer(), 1000),
		),
	).Render(&b, data, document)
	if err != nil {
		t.Fatal(err)
	}

	goldie.New(t).Assert(t, "theme", b.Bytes())
}

func TestStyleInjection(t *testing.T) {
	document := NewParser().Parse(text.NewReader([]byte(`
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
