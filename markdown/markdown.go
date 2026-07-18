package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var gm = goldmark.New(
	// goldmark.WithExtensions(
	// 	meta.Meta,
	// ),
	goldmark.WithRendererOptions(
	// renderer.WithNodeRenderers(
	// 	util.Prioritized(&imageRenderer{
	// 		attachmentProvider: func(source string) (contentID string, err error) {
	// 			return "", errors.New("attachment provided is not implemented")
	// 		},
	// 	}, 500),
	// ),
	),
)

func New() goldmark.Markdown {
	return gm
}

func NewParser() parser.Parser {
	return parser.NewParser(parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(
			parser.DefaultParagraphTransformers()...,
		),
		parser.WithASTTransformers(
			util.Prioritized(&ActionButtonInjector{}, 100),
		),
	)
}

func GetFirstHeadingText(source []byte) (result string) {
	_ = ast.Walk(
		gm.Parser().Parse(text.NewReader(source)),
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
