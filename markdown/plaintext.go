package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// plaintextRenderer renders markdown nodes to plain text without any HTML or HTML comments.
type plaintextRenderer struct{}

// RegisterFuncs registers all node rendering functions.
func (p *plaintextRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// Document and root
	reg.Register(ast.KindDocument, p.renderDocument)
	reg.Register(ast.KindBlockquote, p.renderBlockquote)

	// Headings
	reg.Register(ast.KindHeading, p.renderHeading)

	// Lists
	reg.Register(ast.KindList, p.renderList)
	reg.Register(ast.KindListItem, p.renderListItem)

	// Paragraphs and text
	reg.Register(ast.KindParagraph, p.renderParagraph)
	reg.Register(ast.KindText, p.renderText)
	reg.Register(ast.KindTextBlock, p.renderTextBlock)

	// Code
	reg.Register(ast.KindCodeBlock, p.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, p.renderFencedCodeBlock)
	reg.Register(ast.KindCodeSpan, p.renderCodeSpan)

	// Emphasis
	reg.Register(ast.KindEmphasis, p.renderEmphasis)

	// Links and images
	reg.Register(ast.KindLink, p.renderLink)
	reg.Register(ast.KindImage, p.renderImage)
	reg.Register(ast.KindAutoLink, p.renderAutoLink)

	// Other inline elements
	reg.Register(ast.KindThematicBreak, p.renderThematicBreak)

	// HTML - skip rendering
	reg.Register(ast.KindHTMLBlock, p.renderHTMLBlock)
	reg.Register(ast.KindRawHTML, p.renderRawHTML)

}

func (p *plaintextRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// For documents, we just continue traversing children
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// Blockquotes are rendered with indentation
		_, _ = w.WriteString("> ")
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	h := node.(*ast.Heading)
	if entering {
		// Add newline before heading if not at start
		_ = w.WriteByte('\n')
		// Render heading level with # symbols
		switch h.Level {
		case 1, 2:
		default:
			for i := 0; i < h.Level; i++ {
				_ = w.WriteByte('#')
			}
			_, _ = w.WriteString(" ")
		}
	} else {
		switch h.Level {
		case 1:
			_, _ = w.WriteString("\n")
			lineLength := 40
			_ = ast.Walk(node.FirstChild(), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				text, ok := n.(*ast.Text)
				if ok {
					lineLength = max(len(text.Value(source)), 5)
					return ast.WalkStop, nil
				}
				return ast.WalkContinue, nil
			})
			for i := 0; i < lineLength; i++ {
				_ = w.WriteByte('=')
			}
		case 2:
			_, _ = w.WriteString("\n")
			lineLength := 40
			_ = ast.Walk(node.FirstChild(), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				text, ok := n.(*ast.Text)
				if ok {
					lineLength = max(len(text.Value(source)), 5)
					return ast.WalkStop, nil
				}
				return ast.WalkContinue, nil
			})
			for i := 0; i < lineLength; i++ {
				_ = w.WriteByte('-')
			}
		}
		_, _ = w.WriteString("\n\n")
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_ = w.WriteByte('\n')
	} else {
		_ = w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// Determine if this is part of an ordered or unordered list
		parent := node.Parent()
		if l, ok := parent.(*ast.List); ok {
			if l.IsOrdered() {
				// Simple numbering (doesn't track actual list number, just uses index)
				var index int
				for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
					index++
					if child == node {
						break
					}
				}
				_, _ = w.WriteString(itoa(index))
				_, _ = w.WriteString(". ")
			} else {
				_, _ = w.WriteString("- ")
			}
		}
	} else {
		_ = w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		_, _ = w.WriteString("\n\n")
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		t := node.(*ast.Text)
		_, _ = w.Write(t.Value(source))
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		tb := node.(*ast.TextBlock)
		lines := tb.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			_, _ = w.Write(line.Value(source))
		}
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("\n")
		lines := node.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			_, _ = w.Write(line.Value(source))
		}
		_, _ = w.WriteString("\n")
	}
	return ast.WalkSkipChildren, nil
}

func (p *plaintextRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("\n")
		lines := node.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			_, _ = w.Write(line.Value(source))
		}
		_, _ = w.WriteString("\n")
	}
	return ast.WalkSkipChildren, nil
}

func (p *plaintextRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		cs := node.(*ast.CodeSpan)
		_ = w.WriteByte('`')
		text := cs.FirstChild().(*ast.Text)
		_, _ = w.Write(text.Value(source))
		_ = w.WriteByte('`')
	}
	return ast.WalkSkipChildren, nil
}

func (p *plaintextRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	em := node.(*ast.Emphasis)
	if entering {
		if em.Level == 1 {
			_ = w.WriteByte('*')
		} else if em.Level == 2 {
			_, _ = w.WriteString("**")
		}
	} else {
		if em.Level == 1 {
			_ = w.WriteByte('*')
		} else if em.Level == 2 {
			_, _ = w.WriteString("**")
		}
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	l := node.(*ast.Link)
	if entering {
		_ = w.WriteByte('[')
	} else {
		_, _ = w.WriteString("](")
		_, _ = w.Write(l.Destination)
		_ = w.WriteByte(')')
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	img := node.(*ast.Image)
	if entering {
		_, _ = w.WriteString("![")
	} else {
		_, _ = w.WriteString("](")
		_, _ = w.Write(img.Destination)
		_ = w.WriteByte(')')
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		al := node.(*ast.AutoLink)
		_, _ = w.Write(al.URL(source))
	}
	return ast.WalkContinue, nil
}

func (p *plaintextRenderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("\n---\n")
	}
	return ast.WalkSkipChildren, nil
}

func (p *plaintextRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Skip HTML blocks entirely
	return ast.WalkSkipChildren, nil
}

func (p *plaintextRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Skip raw HTML entirely
	return ast.WalkSkipChildren, nil
}

// itoa converts an integer to a string.
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}

// NewPlaintextRenderer creates and returns a new PlaintextRenderer instance.
func NewPlaintextRenderer() renderer.Renderer {
	return renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(&plaintextRenderer{}, 1000)))
}
