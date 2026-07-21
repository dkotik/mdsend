package template

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/address"
	"github.com/dkotik/mdsend/header"
	"github.com/yuin/goldmark/text"
)

var buffers = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func (t *tmpl) RenderLetterForRecipient(recipient map[string]any) (m mdsend.Message, err error) {
	switch name := recipient[address.FieldName].(type) {
	case string:
		m.To.Name = strings.TrimSpace(name)
	default:
		return m, fmt.Errorf("unexpected type for recipient name: %+v (%T)", name, name)
	}
	switch email := recipient[address.FieldEmail].(type) {
	case string:
		email = strings.TrimSpace(email)
		if email == "" {
			return m, fmt.Errorf("recipient address is empty")
		}
		if err = address.ValidateFormat(email); err != nil {
			return m, fmt.Errorf("recipient address is invalid: %w", err)
		}
		m.To.Address = email
	default:
		return m, fmt.Errorf("unexpected type for recipient email: %+v (%T)", email, email)
	}

	t.Reset()
	b := buffers.Get().(*bytes.Buffer)
	defer func(b *bytes.Buffer) {
		b.Reset()
		buffers.Put(b)
	}(b)
	// t.mu.Lock()
	// defer t.mu.Unlock()
	t.context.Recipient = recipient
	t.context.Content = template.HTML("") // reset

	if err = t.Subject.Execute(b, t.context); err != nil {
		// TODO: use rendering error for all errors in this func
		return m, fmt.Errorf("unable to render subject: %w", err)
	}
	m.Subject = b.String()
	t.context.Frontmatter[mdsend.FieldNameSubject] = m.Subject
	b.Reset()

	for _, h := range t.Headers {
		if err = h.Template.Execute(b, t.context); err != nil {
			return m, fmt.Errorf("unable to render header %q: %w", h.Name, err)
		}
		if b.Len() == 0 {
			continue // skip empty headers
		}
		m.Headers = append(m.Headers, header.Header{
			Name:  h.Name,
			Value: b.String(),
		})
		b.Reset()
	}

	sourceBuffer := buffers.Get().(*bytes.Buffer)
	defer func(b *bytes.Buffer) {
		b.Reset()
		buffers.Put(b)
	}(sourceBuffer)
	t.context.IsPlainText = true
	if err = t.Text.Execute(sourceBuffer, t.context); err != nil {
		return m, fmt.Errorf("unable to render text template: %w", err)
	}
	source := sourceBuffer.Bytes()
	tree := t.ContentParser.Parse(text.NewReader(source))
	if err = t.RendererForText.Render(b, source, tree); err != nil {
		return m, fmt.Errorf("unable to render text: %w", err)
	}
	m.Text = b.String()
	b.Reset()
	sourceBuffer.Reset()

	t.context.IsPlainText = false
	if err = t.Text.Execute(sourceBuffer, t.context); err != nil {
		return m, fmt.Errorf("unable to render text template: %w", err)
	}
	source = sourceBuffer.Bytes()
	tree = t.ContentParser.Parse(text.NewReader(source))
	if err = t.RendererForHTML.Render(b, source, tree); err != nil {
		return m, fmt.Errorf("unable to render HTML: %w", err)
	}
	t.context.Content = template.HTML(b.String())
	b.Reset()
	if err = t.HTML.Execute(b, t.context); err != nil {
		return m, fmt.Errorf("unable to render HTML template: %w", err)
	}
	m.HTML = b.String()
	m.LetterID = t.LetterID
	m.ID, err = t.IdentifierGenerator.GenerateID()
	if err != nil {
		return m, err
	}
	m.SeedKey = t.SeedKey
	m.From = t.From
	m.ScheduleAfter = t.context.Schedule.After
	t.context.Schedule = t.context.Schedule.Next()
	return m, nil
}
