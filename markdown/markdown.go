package markdown

import (
	"errors"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func New() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(
				util.Prioritized(&imageRenderer{
					attachmentProvider: func(source string) (contentID string, err error) {
						return "", errors.New("attachment provided is not implemented")
					},
				}, 500),
			),
		),
	)
}
