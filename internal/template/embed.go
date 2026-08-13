package template

import (
	"embed"
	"errors"
	"io/fs"
	"path"
	"strings"
	"sync"
)

var (
	//go:embed html/*
	defaultTemplates embed.FS

	defaultTemplateSyncOnce = &sync.Once{}
	defaultTemplate         []byte
)

func getDefaultTemplateHTML() []byte {
	defaultTemplateSyncOnce.Do(func() {
		var err error
		defaultTemplate, err = defaultTemplates.ReadFile("html/default.html")
		if err != nil {
			panic(err)
		}
	})
	return defaultTemplate
}

type fsWithEmbeddedTemplates struct {
	Wrapped fs.FS
}

func NewFileSystemWithEmbeddedTemplates(wrap fs.FS) fs.FS {
	if wrap == nil {
		panic("nil template file system")
	}
	return fsWithEmbeddedTemplates{Wrapped: wrap}
}

func (f fsWithEmbeddedTemplates) Open(p string) (fs.File, error) {
	file, err := f.Wrapped.Open(p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) && strings.HasPrefix(p, "mdsend://") {
			p = strings.TrimPrefix(p, "mdsend://")
			p = path.Join("html", p)
			return defaultTemplates.Open(p)
		}
		return nil, err
	}
	return file, nil
}
