package template

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/mail"
	"strings"
	ttemplate "text/template"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal"
	"github.com/dkotik/mdsend/markdown"
	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
)

//go:embed html/*
var defaultTemplates embed.FS

type Template interface {
	RenderLetterForRecipient(map[string]any) (mdsend.Message, error)
}

type IdentifierGenerator interface {
	GenerateID() (string, error)
}

type IdentifierGeneratorFunc func() (string, error)

func (f IdentifierGeneratorFunc) GenerateID() (string, error) {
	return f()
}

type tmpl struct {
	LetterID            string
	SeedKey             string
	From                mail.Address
	Headers             []headerTemplate
	Subject             *ttemplate.Template
	Text                *ttemplate.Template
	ReifiedCache        map[string]string
	HTML                *template.Template
	ContentParser       parser.Parser
	RendererForText     renderer.Renderer
	RendererForHTML     renderer.Renderer
	IdentifierGenerator IdentifierGenerator

	// mu      *sync.Mutex
	context Context
}

type Options struct {
	IdentifierGenerator IdentifierGenerator
	ContentParser       parser.Parser
	RendererForText     renderer.Renderer
	RendererForHTML     renderer.Renderer
	Frontmatter         map[string]any
}

func (o Options) withDefaults() Options {
	if o.IdentifierGenerator == nil {
		o.IdentifierGenerator = IdentifierGeneratorFunc(func() (string, error) {
			return uuid.NewString(), nil
		})
	}
	if o.ContentParser == nil {
		o.ContentParser = goldmark.DefaultParser()
	}
	if o.RendererForText == nil {
		o.RendererForText = markdown.NewPlaintextRenderer()
	}
	if o.RendererForHTML == nil {
		o.RendererForHTML = markdown.New().Renderer()
	}
	if o.Frontmatter == nil {
		o.Frontmatter = make(map[string]any)
	}
	return o
}

// New creates a [Template]. The result is not safe for asynchronous rendering.
func New(
	l mdsend.Letter,
	options Options,
) (_ Template, err error) {
	options = options.withDefaults()
	internal.MapMergeLeft(options.Frontmatter, l.Frontmatter)
	t := &tmpl{
		LetterID: l.ID,
		// mu:              &sync.Mutex{},
		ContentParser:       options.ContentParser,
		RendererForText:     options.RendererForText,
		RendererForHTML:     options.RendererForHTML,
		IdentifierGenerator: options.IdentifierGenerator,
		context: Context{
			Frontmatter: options.Frontmatter,
			Content:     template.HTML(l.Content), // for initial templates only
		},
	}
	if t.context.Frontmatter == nil {
		t.context.Frontmatter = make(map[string]any)
	}
	t.context.Schedule, err = l.GetSchedule()
	if err != nil {
		return nil, err
	}
	t.SeedKey, err = newSeedKey(t.context, l)
	if err != nil {
		return nil, err
	}
	t.From, err = l.GetFrom()
	if err != nil {
		return nil, err
	}

	templateFunctions := functions()
	templateFunctions["reify"] = t.Reify
	l.Content = strings.TrimSpace(l.Content)
	if l.Content == "" {
		return nil, errors.New("empty letter content")
	}
	t.Text, err = ttemplate.New("").Funcs(templateFunctions).Parse(l.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to parse letter content as a template: %w", err)
	}

	subject, err := l.GetSubject()
	if err != nil {
		return nil, err
	}
	t.Subject, err = t.Text.New("").Parse(subject)
	if err != nil {
		return nil, fmt.Errorf("invalid subject template: %w", err)
	}
	headers, err := l.GetHeaders()
	if err != nil {
		return nil, err
	}
	t.Headers = make([]headerTemplate, len(headers))
	for i, header := range headers {
		value, err := t.Text.New("").Parse(header.Value)
		if err != nil {
			return nil, fmt.Errorf("unable to parse header %q as template: %w", header.Name, err)
		}
		t.Headers[i] = headerTemplate{
			Name:     header.Name,
			Template: value,
		}
	}

	t.HTML = template.New("").Funcs(templateFunctions)
	for _, subTemplate := range l.Templates {
		t.HTML, err = t.HTML.Parse(string(subTemplate.Content))
		if err != nil {
			return nil, fmt.Errorf("unable to parse template %q: %w", subTemplate.Name, err)
		}
	}

	if len(l.Templates) == 0 {
		defaultTemplate, err := defaultTemplates.ReadFile("html/default.html")
		if err != nil {
			return nil, fmt.Errorf("unable to load default template: %w", err)
		}
		t.HTML, err = t.HTML.Parse(string(defaultTemplate))
		if err != nil {
			return nil, fmt.Errorf("unable to parse default template: %w", err)
		}
	}
	return t, nil
}
