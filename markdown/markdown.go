package markdown

import (
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
)

var gm = goldmark.New(
	goldmark.WithExtensions(
		meta.Meta,
	),
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
