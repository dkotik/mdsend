package template

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/markdown"
	"github.com/yuin/goldmark"
)

type Template interface {
	Render(plainText, HTML io.Writer, recipient map[string]any) error
}

func Parse(text string) (*template.Template, error) {
	return template.New("").Funcs(Functions()).Parse(text)
}

type messageTemplate struct {
	MarkdownRenderer goldmark.Markdown
	Frontmatter      map[string]any
	Template         *template.Template
	Content          *template.Template

	// mu *sync.Mutex
	Text *bytes.Buffer
	HTML *bytes.Buffer
}

// New creates a [Template]. It is not safe for asynchronous
// rendering.
func New(
	m mdsend.Letter,
	r goldmark.Markdown,
) (Template, error) {
	if r == nil {
		return nil, errors.New("renderer for Markdown is nil")
	}
	content, err := Parse(m.Content)
	if err != nil {
		return messageTemplate{}, fmt.Errorf("unable to parse template fields in message content: %w", err)
	}
	raw, err := loadTemplate(m)
	if err != nil {
		return messageTemplate{}, fmt.Errorf("unable to load message template: %w", err)
	}
	template, err := Parse(string(raw))
	if err != nil {
		return messageTemplate{}, fmt.Errorf("unable to parse message template: %w", err)
	}
	frontmatter := m.Frontmatter
	if frontmatter == nil {
		frontmatter = make(map[string]any)
	}
	return messageTemplate{
		MarkdownRenderer: r,
		Frontmatter:      frontmatter,
		Template:         template,
		Content:          content,

		Text: &bytes.Buffer{},
		HTML: &bytes.Buffer{},
	}, nil
}

func (t messageTemplate) Render(
	plainText, HTML io.Writer,
	recipient map[string]any,
) (err error) {
	context := Context{
		Frontmatter: t.Frontmatter,
		Recipient:   recipient,
	}

	t.Text.Reset()
	if err = t.Content.Execute(
		io.MultiWriter(plainText, t.Text),
		context,
	); err != nil {
		return fmt.Errorf("unable to execute content template: %w", err)
	}

	t.HTML.Reset()
	if err = markdown.New().Convert(
		t.Text.Bytes(),
		t.HTML,
	); err != nil {
		return err
	}
	context.Content = template.HTML(t.HTML.String())

	if err = t.Template.Execute(HTML, context); err != nil {
		return fmt.Errorf("unable to execute message template: %w", err)
	}
	return nil
}
