package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type OptionsHTML struct {
	BlockQuoteStyle      string
	LinkStyle            string
	ActionContainerStyle string
	ActionStyle          string
}

func (options OptionsHTML) withDefaults() OptionsHTML {
	if options.BlockQuoteStyle == "" {
		options.BlockQuoteStyle = "border-left: 4px solid #ccc; margin: 15px 0; padding: 15px 20px; background-color: #f9f9f9; color: #555;"
	}
	if options.LinkStyle == "" {
		options.LinkStyle = "color: #0066cc; text-decoration: underline;"
	}
	if options.ActionContainerStyle == "" {
		options.ActionContainerStyle = "border-radius: 5px; background-color:#3a86ff;"
	}
	if options.ActionStyle == "" {
		options.ActionStyle = "font-size: 18px; color: #ffffff; font-weight: bold; text-decoration: none;border-radius: 5px; padding: 12px 18px; border: 1px solid #3a86ff; display: inline-block;"
	}
	return options
}

type htmlRenderer struct {
	BlockQuoteStyle      string
	LinkStyle            string
	ActionContainerStyle string
	ActionStyle          string
	Writer               html.Writer
}

func NewRendererHTML(options OptionsHTML) (renderer.Renderer, error) {
	options = options.withDefaults()
	return renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(&htmlRenderer{
				BlockQuoteStyle:      options.BlockQuoteStyle,
				LinkStyle:            options.LinkStyle,
				ActionContainerStyle: options.ActionContainerStyle,
				ActionStyle:          options.ActionStyle,
				Writer:               html.DefaultWriter,
			}, 1000),
			util.Prioritized(html.NewRenderer(), 1000),
		),
	), nil
}

func (r *htmlRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(KindActionButton, r.renderActioButtonNodeHTML)
}

func (r *htmlRenderer) renderBlockquote(
	w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<blockquote style=\"")
			_, _ = w.WriteString(r.BlockQuoteStyle)
			_ = w.WriteByte('>')
			html.RenderAttributes(w, n, html.BlockquoteAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<blockquote>\n")
		}
	} else {
		_, _ = w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}

func (r *htmlRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("<a href=\"")
		dest := util.URLEscape(n.Destination, true)
		// if r.Unsafe || !html.IsDangerousURL(dest) {
		_, _ = w.Write(util.EscapeHTML(dest))
		// }
		_ = w.WriteByte('"')
		if n.Title != nil {
			_, _ = w.WriteString(` title="`)
			r.Writer.Write(w, n.Title)
			_ = w.WriteByte('"')
		}
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.LinkAttributeFilter)
		}
		_, _ = w.WriteString(` style="`)
		_, _ = w.WriteString(r.LinkStyle)
		_, _ = w.WriteString(`">`)
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

func (r *htmlRenderer) renderActioButtonNodeHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ActionButtonNode)
	if entering {
		_, _ = w.WriteString(`
		<table border="0" cellspacing="0" cellpadding="0">
			<tr>
        <td align="center" style="`)
		_, _ = w.WriteString(r.ActionContainerStyle)
		_, _ = w.WriteString(`">
		<table border="0" cellspacing="0" cellpadding="0">
      <tr>
          <td align="center" style="border-radius: 5px; background-color:#3a86ff;">
              <a rel="noopener" target="_blank" href="`)
		_, _ = w.Write(util.URLEscape(n.Link.Destination, false))
		_, _ = w.WriteString(`" title="`)
		r.Writer.Write(w, n.Link.Title)
		_, _ = w.WriteString(`" style="`)
		_, _ = w.WriteString(r.ActionStyle)
		_, _ = w.WriteString(`">`)
	} else {
		_, _ = w.WriteString(` &rarr;</a>
                      </td>
                  </tr>
              </table></td></tr></table>`)
	}
	return ast.WalkContinue, nil
}
