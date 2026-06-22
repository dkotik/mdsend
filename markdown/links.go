package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var attachmentsSniffer = parser.NewParser(
	parser.WithBlockParsers(
		util.Prioritized(parser.NewSetextHeadingParser(), 100),
		util.Prioritized(parser.NewThematicBreakParser(), 200),
		util.Prioritized(parser.NewListParser(), 300),
		util.Prioritized(parser.NewListItemParser(), 400),
		util.Prioritized(parser.NewCodeBlockParser(), 500),
		util.Prioritized(parser.NewATXHeadingParser(), 600),
		util.Prioritized(parser.NewFencedCodeBlockParser(), 700),
		util.Prioritized(parser.NewBlockquoteParser(), 800),
		// util.Prioritized(parser.NewHTMLBlockParser(), 900),
		util.Prioritized(parser.NewParagraphParser(), 1000),
	),
	parser.WithInlineParsers(
		util.Prioritized(parser.NewCodeSpanParser(), 100),
		util.Prioritized(parser.NewLinkParser(), 200),
		util.Prioritized(parser.NewAutoLinkParser(), 300),
		util.Prioritized(parser.NewRawHTMLParser(), 400),
		util.Prioritized(parser.NewEmphasisParser(), 500),
	),
	parser.WithParagraphTransformers(
		util.Prioritized(parser.LinkReferenceParagraphTransformer, 100),
	),
)

type Link struct {
	Name string
	Path string
}

func CollectLinks(source []byte) (result []Link) {
	pc := parser.NewContext()
	ast.Walk(
		attachmentsSniffer.Parse(
			text.NewReader(source),
			parser.WithContext(pc),
		),
		ast.Walker(func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				switch n := n.(type) {
				case *ast.Link:
					result = append(result, Link{Name: string(n.Title), Path: string(n.Destination)})
				case *ast.Image:
					result = append(result, Link{Name: string(n.Title), Path: string(n.Destination)})
				}
			}
			return ast.WalkContinue, nil
		}),
	)
	return result
}
