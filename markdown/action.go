package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindActionButton = ast.NewNodeKind("ActionButton")

type ActionButtonNode struct {
	ast.Link
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
	// ast.Walk traverses every node in the document tree
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		// Entering is true when arriving at a node, false when leaving it
		if !entering {
			return ast.WalkContinue, nil
		}
		p, ok := n.(*ast.Paragraph)
		if !ok || p.ChildCount() != 1 {
			return ast.WalkContinue, nil
		}
		// panic("dd")
		link, ok := p.FirstChild().(*ast.Link)
		if !ok || link == nil {
			return ast.WalkContinue, nil
		}
		p.ReplaceChild(p, link, &ActionButtonNode{
			Link: *link,
		})
		// n.Parent().ReplaceChild(n.Parent(), n, &ActionButtonNode{
		// 	Link: *link,
		// })

		return ast.WalkContinue, nil
	})
}

type actionRenderer struct{}

func (r *actionRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindActionButton, r.renderActioButtonNodeHTML)
}

func (r *actionRenderer) renderActioButtonNodeHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ActionButtonNode)
	if entering {
		_, _ = w.WriteString(`<table border="0" cellspacing="0" cellpadding="0">
      <tr>
          <td align="center" style="border-radius: 5px; background-color:#3a86ff;">
              <a rel="noopener" target="_blank" target="_blank" href="`)
		_, _ = w.Write(util.URLEscape(n.Destination, false))
		_, _ = w.WriteString(`" title="`)
		_, _ = w.Write(util.EscapeHTML(n.Title))
		_, _ = w.WriteString(`" target="_blank" style="font-size: 18px; color: #ffffff; font-weight: bold; text-decoration: none;border-radius: 5px; padding: 12px 18px; border: 1px solid #3a86ff; display: inline-block;">
			`)
	} else {
		_, _ = w.WriteString(` &rarr;</a>
                      </td>
                  </tr>
              </table>`)
	}
	return ast.WalkContinue, nil
}
