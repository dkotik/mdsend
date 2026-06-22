package loader

import (
	"context"
	"io/fs"

	"github.com/dkotik/mdsend"
)

type Loader interface {
	LoadLetter(context.Context, string) (mdsend.Letter, Recipients, error)
}

type loader struct {
	FS       fs.FS
	Markdown MarkdownConfiguration

	Cache Cache
	// TemplateCache map[string]*template.Template
}

func NewLoader(config Configuration) (_ Loader, _ Recipients, err error) {
	if config, err = config.withDefaults(); err != nil {
		return nil, nil, err
	}
	return loader{
		FS:       config.FS,
		Markdown: config.Markdown,
		Cache:    NewMapCache(),
		// TemplateCache: make(map[string]*template.Template),
	}, nil, nil
}

func (l loader) getFile(ctx context.Context, p string) ([]byte, error) {
	data, err := l.Cache.Pull(ctx, p)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return data, nil
	}
	data, err = fs.ReadFile(l.FS, p)
	if err != nil {
		return nil, err
	}
	if err := l.Cache.Push(ctx, p, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (l loader) loadLetter(ctx context.Context, p string) (mdsend.Letter, error) {
	data, err := l.getFile(ctx, p)
	if err != nil {
		return mdsend.Letter{}, err
	}
	return mdsend.NewLetter(data)
}

func (l loader) LoadLetter(ctx context.Context, p string) (mdsend.Letter, Recipients, error) {
	letter, err := l.loadLetter(ctx, p)
	if err != nil {
		return mdsend.Letter{}, nil, err
	}
	return letter, nil, nil
}
