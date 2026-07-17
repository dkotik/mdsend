package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
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
