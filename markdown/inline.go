package markdown

import (
	"bytes"
	"iter"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	reImageDestination     = regexp.MustCompile(`^\!\[[^\]]*\]\(\s*(\S+)`)
	reReferenceDestination = regexp.MustCompile(`^\[[^\]]+\]\:\s*(\S+)`)
)

type InlineImage struct {
	Title string
	Alt   string
	// Reference     string
	Location   string
	LocationAt int
}

// IndexImages finds all inline images and references
// that might point to an inline image in the source content.
func IndexImages(parser parser.Parser, source []byte) iter.Seq[InlineImage] {
	return func(yield func(InlineImage) bool) {
		tree := parser.Parse(text.NewReader(source))
		references := []*ast.LinkReferenceDefinition{}
		images := []*ast.Image{}
		ast.Walk(tree, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			if image, ok := node.(*ast.Image); ok {
				images = append(images, image)
			}
			if reference, ok := node.(*ast.LinkReferenceDefinition); ok {
				references = append(references, reference)
			}
			return ast.WalkContinue, nil
		})

		for _, image := range images {
			if image.Reference != nil {
				for _, reference := range references {
					if bytes.Equal(reference.Label, image.Reference.Value) {
						at := reference.Pos()
						m := reReferenceDestination.FindSubmatchIndex(source[at:])
						if m == nil {
							continue
						}
						if !yield(InlineImage{
							Title:      string(reference.Title),
							Alt:        getImageAltText(source, image),
							Location:   string(reference.Destination),
							LocationAt: at + m[2],
						}) {
							return
						}
						break
					}
				}
				continue
			}
			at := image.Pos()
			m := reImageDestination.FindSubmatchIndex(source[at:])
			if m == nil {
				continue
			}
			if !yield(InlineImage{
				Title:      string(image.Title),
				Alt:        getImageAltText(source, image),
				Location:   string(image.Destination),
				LocationAt: at + m[2],
			}) {
				return
			}
		}
	}
}
