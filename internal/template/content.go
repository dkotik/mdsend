package template

import (
	"bytes"
	"io"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type executable interface {
	Execute(w io.Writer, data any) error
}

func renderTemplate(t executable, data any) (_ string, err error) {
	b := buffers.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		buffers.Put(b)
	}()
	if err = t.Execute(b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

func renderMarkdown(
	r renderer.Renderer,
	source []byte,
	tree ast.Node,
) (_ string, err error) {
	b := buffers.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		buffers.Put(b)
	}()
	if err = r.Render(b, source, tree); err != nil {
		return "", err
	}
	return b.String(), nil
}
