package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func NewParser(theme Theme) parser.Parser {
	return parser.NewParser(parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(
			parser.DefaultParagraphTransformers()...,
		),
		parser.WithASTTransformers(
			util.Prioritized(&ActionButtonInjector{}, 100),
			util.Prioritized(theme, 1000),
		),
	)
}

func NewRendererHTML() renderer.Renderer {
	return renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(&ActionButtonInjector{}, 100),
			util.Prioritized(html.NewRenderer(), 1000),
		),
	)
}

func GetFirstHeadingText(source []byte) (result string) {
	_ = ast.Walk(
		goldmark.DefaultParser().Parse(text.NewReader(source)),
		ast.Walker(func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering && n.Kind() == ast.KindHeading {
				result = string(bytes.TrimSpace(
					n.Lines().Value(source),
				))
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		}),
	)
	return result
}

func moveAllChildrenTo(target, source ast.Node) {
	// 1. Collect all children
	var children []ast.Node
	for child := source.FirstChild(); child != nil; child = child.NextSibling() {
		children = append(children, child)
	}

	// 2. Remove from source
	source.RemoveChildren(source)

	// 3. Append to target and update parent
	for _, child := range children {
		child.SetParent(target)
		target.AppendChild(target, child)
	}
}

func moveAllSiblingsTo(target, source ast.Node) {
	// 1. Collect all siblings in a slice so we can safely modify them
	var siblings []ast.Node
	for current := source.NextSibling(); current != nil; current = current.NextSibling() {
		siblings = append(siblings, current)
	}

	// 2. Remove each sibling from its current parent and add it to the new target
	for _, sibling := range siblings {
		sibling.Parent().RemoveChild(sibling.Parent(), sibling)
		target.AppendChild(target, sibling)
	}
}

func getListDepth(n ast.Node) int {
	depth := 0
	for p := n.Parent(); p != nil; p = p.Parent() {
		if p.Kind() == ast.KindList {
			depth++
		}
	}
	return depth - 1
}
