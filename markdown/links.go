package markdown

import (
	"bytes"
	"io"
	"path"

	"github.com/dkotik/mdsend/internal/media"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var attachmentsSniffer = parser.NewParser(
	parser.WithBlockParsers(
		// util.Prioritized(parser.NewSetextHeadingParser(), 100),
		// util.Prioritized(parser.NewThematicBreakParser(), 200),
		util.Prioritized(parser.NewListParser(), 300),
		util.Prioritized(parser.NewListItemParser(), 400),
		util.Prioritized(parser.NewCodeBlockParser(), 500),
		// util.Prioritized(parser.NewATXHeadingParser(), 600),
		util.Prioritized(parser.NewFencedCodeBlockParser(), 700),
		// util.Prioritized(parser.NewBlockquoteParser(), 800),
		// util.Prioritized(parser.NewHTMLBlockParser(), 900),
		util.Prioritized(parser.NewParagraphParser(), 1000),
	),
	parser.WithInlineParsers(
		util.Prioritized(parser.NewCodeSpanParser(), 100),
		util.Prioritized(parser.NewLinkParser(), 200),
		util.Prioritized(parser.NewAutoLinkParser(), 300),
		util.Prioritized(parser.NewRawHTMLParser(), 400),
		// util.Prioritized(parser.NewEmphasisParser(), 500),
	),
	parser.WithParagraphTransformers(
		// seems to be essential
		util.Prioritized(parser.LinkReferenceParagraphTransformer, 100),
	),
)

type Link struct {
	Name        string
	Destination string
	// Position    int
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
					if n.Reference != nil {
						return ast.WalkContinue, nil
					}
					// index := n.Pos() + 3
					// if n.FirstChild() != nil {
					// 	index += len(n.FirstChild().(*ast.Text).Value(source))
					// }
					result = append(result, Link{
						Name:        string(n.Title),
						Destination: string(n.Destination),
						// Position:    index,
					})
					// if n.NextSibling() != nil {
					// 	result[len(result)-1].Position = n.NextSibling().Pos() + 1
					// }
				case *ast.Image:
					if n.Reference != nil {
						return ast.WalkContinue, nil
					}
					// index := n.Pos() + 4
					// if n.FirstChild() != nil {
					// 	index += len(n.FirstChild().(*ast.Text).Value(source))
					// }
					result = append(result, Link{
						Name:        string(n.Title),
						Destination: string(n.Destination),
						// Position:    index,
					})
				case *ast.LinkReferenceDefinition:
					// index := n.Pos() + 3
					// if n.FirstChild() != nil {
					// 	index += len(n.FirstChild().(*ast.Text).Value(source))
					// }
					result = append(result, Link{
						Name:        string(n.Title),
						Destination: string(n.Destination),
						// Position:    n.Pos() + 3 + len(n.Label),
					})
				}
			}
			return ast.WalkContinue, nil
		}),
	)
	return result
}

func CopyWithRelativePathPrefix(
	w io.Writer,
	source []byte,
	relativePathPrefix string,
) (err error) {
	written := 0
	err = ast.Walk(
		attachmentsSniffer.Parse(
			text.NewReader(source),
		),
		ast.Walker(func(n ast.Node, entering bool) (_ ast.WalkStatus, err error) {
			if entering {
				switch n := n.(type) {
				case *ast.Link:
					if n.Reference != nil {
						return ast.WalkContinue, nil
					}
					p := bytes.TrimSpace(n.Destination)
					if !media.IsPathLocalBytes(p) {
						return ast.WalkContinue, nil
					}
					index := n.Pos() + 3
					if n.FirstChild() != nil {
						index += len(n.FirstChild().(*ast.Text).Value(source))
					}
					if _, err = w.Write(source[written:index]); err != nil {
						return ast.WalkStop, err
					}
					if _, err = w.Write([]byte(path.Join(
						relativePathPrefix,
						string(p),
					))); err != nil {
						return ast.WalkStop, err
					}
					written = index + len(n.Destination)
				case *ast.Image:
					if n.Reference != nil {
						return ast.WalkContinue, nil
					}
					p := bytes.TrimSpace(n.Destination)
					if !media.IsPathLocalBytes(p) {
						return ast.WalkContinue, nil
					}
					index := n.Pos() + 4
					if n.FirstChild() != nil {
						index += len(n.FirstChild().(*ast.Text).Value(source))
					}
					if _, err = w.Write(source[written:index]); err != nil {
						return ast.WalkStop, err
					}
					if _, err = w.Write([]byte(path.Join(
						relativePathPrefix,
						string(p),
					))); err != nil {
						return ast.WalkStop, err
					}
					written = index + len(n.Destination)
				case *ast.LinkReferenceDefinition:
					p := bytes.TrimSpace(n.Destination)
					if !media.IsPathLocalBytes(p) {
						return ast.WalkContinue, nil
					}
					index := n.Pos() + 3 + len(n.Label)

				drainWhiteSpace:
					for ; index < len(source); index++ {
						switch source[index] {
						case ' ', '\t', '\n', '\r':
							continue
						default:
							break drainWhiteSpace
						}
					}
					if _, err = w.Write(source[written:index]); err != nil {
						return ast.WalkStop, err
					}
					if _, err = w.Write([]byte(path.Join(
						relativePathPrefix,
						string(p),
					))); err != nil {
						return ast.WalkStop, err
					}
					written = index + len(n.Destination)
				}
			}
			return ast.WalkContinue, nil
		}),
	)
	if err != nil {
		return err
	}
	_, err = w.Write(source[written:])
	return err
}
