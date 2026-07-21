package html

import (
	"embed"
	"errors"
	"os"
	"path"
	"strings"
	"sync"
)

var (
	//go:embed templates/*
	defaultTemplates embed.FS

	defaultTemplateSyncOnce = &sync.Once{}
	defaultTemplate         []byte
)

func LoadEmbeddedTemplate(p string) ([]byte, bool) {
	if !strings.HasPrefix(p, "mdsend://") {
		return nil, false
	}
	p = strings.TrimPrefix(p, "mdsend://")
	p = path.Join("templates", p)
	data, err := defaultTemplates.ReadFile(p)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
		return nil, false
	}
	return data, true
}

func GetDefaultTemplateHTML() []byte {
	defaultTemplateSyncOnce.Do(func() {
		var ok bool
		defaultTemplate, ok = LoadEmbeddedTemplate("mdsend://default.html")
		if !ok {
			panic("default template not found")
		}
	})
	return defaultTemplate
}
