package markdown

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type imageRenderer struct {
}

func (i *imageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, i.renderImage)
}

func (i *imageRenderer) renderImage(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		i := n.(*ast.Image)
		// TODO: inject the template value here?
		// spew.Fdump(w, n)
		// alt := i.FirstChild().(*ast.Text)
		fmt.Fprintf(w, "{ Title: %s, Destination: %s, Alt: %s }", i.Title, i.Destination, i.FirstChild().Text(source))
	}

	return ast.WalkSkipChildren, nil
}
