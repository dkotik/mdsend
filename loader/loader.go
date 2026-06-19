package loader

import (
	"io/fs"

	"github.com/dkotik/mdsend"
)

type Loader interface {
	LoadLetter(string) (mdsend.Letter, Recipients, error)
}

type loader struct {
	FS       fs.FS
	Markdown MarkdownConfiguration

	Cache map[string][]byte
	// TemplateCache map[string]*template.Template
}

func NewLoader(config Configuration) (_ Loader, _ Recipients, err error) {
	if config, err = config.withDefaults(); err != nil {
		return nil, nil, err
	}
	return loader{
		FS:       config.FS,
		Markdown: config.Markdown,
		Cache:    make(map[string][]byte),
		// TemplateCache: make(map[string]*template.Template),
	}, nil, nil
}

func (l loader) load(p string) ([]byte, error) {
	data, ok := l.Cache[p]
	if ok {
		return data, nil
	}
	data, err := fs.ReadFile(l.FS, p)
	if err != nil {
		return nil, err
	}
	l.Cache[p] = data
	return data, nil
}

func (l loader) LoadLetter(p string) (mdsend.Letter, Recipients, error) {
	// data, err := fs.ReadFile(l.FS, p)
	// if err != nil {
	// 	return mdsend.Letter{}, nil, err
	// }

	return mdsend.Letter{}, nil, nil
}
