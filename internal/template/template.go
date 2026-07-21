package template

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/mail"
	"path"
	"strings"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal"
	"github.com/dkotik/mdsend/markdown"
	"github.com/google/uuid"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
)

//go:embed html/*
var defaultTemplates embed.FS

type Template interface {
	RenderLetterForRecipient(map[string]any) (mdsend.Message, error)
}

type tmpl struct {
	LetterID            string
	SeedKey             string
	From                mail.Address
	Headers             []headerTemplate
	Subject             *template.Template
	Text                *template.Template
	HTML                *template.Template
	ReifiedCache        map[string]template.HTML
	ContentParser       parser.Parser
	RendererForText     renderer.Renderer
	RendererForHTML     renderer.Renderer
	IdentifierGenerator mdsend.IdentifierGenerator

	// mu      *sync.Mutex
	context Context
}

type Options struct {
	IdentifierGenerator mdsend.IdentifierGenerator
	ContentParser       parser.Parser
	RendererForText     renderer.Renderer
	RendererForHTML     renderer.Renderer
	Frontmatter         map[string]any
}

// New creates a [Template]. The result is not safe for asynchronous rendering.
func New(
	l mdsend.Letter,
	options Options,
) (_ Template, err error) {
	l.Content = strings.TrimSpace(l.Content)
	if l.Content == "" {
		return nil, errors.New("empty letter content")
	}
	if options.IdentifierGenerator == nil {
		options.IdentifierGenerator = mdsend.IdentifierGeneratorFunc(func() (string, error) {
			return uuid.NewString(), nil
		})
	}
	if options.ContentParser == nil {
		if themeStack, ok := l.Frontmatter["theme"].(map[string]any); ok {
			theme, err := markdown.NewThemeFromMap(themeStack)
			if err != nil {
				return nil, err
			}
			options.ContentParser = markdown.NewParser(theme)
		} else {
			options.ContentParser = markdown.NewParser(markdown.DefaultLightTheme)
		}
	}
	if options.RendererForText == nil {
		options.RendererForText = markdown.NewPlaintextRenderer()
	}
	if options.RendererForHTML == nil {
		options.RendererForHTML = markdown.NewRendererHTML()
	}
	if options.Frontmatter == nil {
		options.Frontmatter = make(map[string]any)
	}
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
	t.From, err = l.GetFrom()
	if err != nil {
		return nil, err
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

	templateFunctions := functions()
	templateFunctions["reify"] = t.Reify
	// templateFunctions["lookup"] = t.Lookup
	t.HTML, err = template.New("").Funcs(templateFunctions).Parse(l.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to parse letter content as a template: %w", err)
	}
	if len(l.Templates) == 0 {
		defaultTemplate, err := defaultTemplates.ReadFile("html/default.html")
		if err != nil {
			return nil, fmt.Errorf("unable to load default template: %w", err)
		}
		t.Text, err = t.HTML.Clone()
		if err != nil {
			return nil, fmt.Errorf("unable to clone letter content as a template: %w", err)
		}
		t.HTML, err = t.HTML.Parse(string(defaultTemplate))
		if err != nil {
			return nil, fmt.Errorf("unable to parse default template: %w", err)
		}
	} else {
		// clone, err := t.HTML.Clone() // for re-inserting
		// if err != nil {
		// 	return nil, fmt.Errorf("unable to clone letter content as a template: %w", err)
		// }
		var subTemplate mdsend.Attachment
		for _, subTemplate = range l.Templates {
			t.HTML, err = t.HTML.New(
				path.Base(subTemplate.Name),
			).Parse(string(subTemplate.Content))
			if err != nil {
				return nil, fmt.Errorf("unable to parse template %q: %w", subTemplate.Name, err)
			}
		}
		t.Text, err = t.HTML.Lookup("").Clone()
		if err != nil {
			return nil, fmt.Errorf("unable to clone letter content as a template: %w", err)
		}
		// the latest template becomes the root template
		t.HTML = WithNewRootTemplate(t.HTML, t.HTML.Lookup(path.Base(subTemplate.Name)))
		// reinsert the clone back into text template
		// t.Text, err = t.Text.AddParseTree("", clone.Tree)
		// t.Text, err = t.Text.New("").Parse(l.Content)
		// if err != nil {
		// 	return nil, fmt.Errorf("unable to reinsert clone into text template: %w", err)
		// }
	}

	subject, err := l.GetSubject()
	if err != nil {
		return nil, err
	}
	clone, err := t.Text.Clone()
	if err != nil {
		return nil, fmt.Errorf("unable to clone content template: %w", err)
	}
	t.Subject, err = clone.New("").Parse(subject)
	if err != nil {
		return nil, fmt.Errorf("invalid subject template: %w", err)
	}
	headers, err := l.GetHeaders()
	if err != nil {
		return nil, err
	}
	t.Headers = make([]headerTemplate, len(headers))
	for i, header := range headers {
		clone, err = t.Text.Clone()
		if err != nil {
			return nil, fmt.Errorf("unable to clone content template: %w", err)
		}
		value, err := clone.New("").Parse(header.Value)
		if err != nil {
			return nil, fmt.Errorf("unable to parse header %q as template: %w", header.Name, err)
		}
		t.Headers[i] = headerTemplate{
			Name:     header.Name,
			Template: value,
		}
	}
	return t, nil
}
