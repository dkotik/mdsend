package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var KindActionButton = ast.NewNodeKind("ActionButton")

type ActionButtonNode struct {
	ast.BaseBlock
	Link ast.Link
}

func (n *ActionButtonNode) Kind() ast.NodeKind {
	return KindActionButton
}

func (n *ActionButtonNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"value": "<nil>",
	}, nil)
}

type ActionButtonInjector struct{}

func (g *ActionButtonInjector) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	type actionReplacement struct {
		Paragraph *ast.Paragraph
		Action    *ActionButtonNode
	}

	replacements := make([]actionReplacement, 0)
	for n := node.FirstChild(); n != nil; n = n.NextSibling() {
		p, ok := n.(*ast.Paragraph)
		if !ok || p.ChildCount() != 1 {
			continue
		}
		link, ok := p.FirstChild().(*ast.Link)
		if !ok || link == nil {
			continue
		}
		action := &ActionButtonNode{
			BaseBlock: ast.BaseBlock{},
			Link:      *link,
		}
		// moveAllSiblingsTo(action, n)
		moveAllChildrenTo(action, link)
		replacements = append(replacements, actionReplacement{
			Paragraph: p,
			Action:    action,
		})
	}

	for _, replacement := range replacements {
		parent := replacement.Paragraph.Parent()
		parent.ReplaceChild(
			parent,
			replacement.Paragraph,
			replacement.Action,
		)
	}
	return
}
