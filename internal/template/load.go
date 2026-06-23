package template

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/dkotik/mdsend"
)

//go:embed html/*
var templates embed.FS

func loadTemplate(m mdsend.Letter) (data []byte, err error) {
	p, ok := m.Frontmatter[mdsend.FieldNameTemplates]
	defer func() {
		if err == nil {
			data = bytes.TrimSpace(data)
			if len(data) == 0 {
				err = errors.New("empty template")
			}
		}
	}()
	if ok {
		switch p := p.(type) {
		case nil:
		case string:
			p = strings.TrimSpace(p)
			if p != "" {
				data, err = os.ReadFile(
					p,
					// filepath.Join(
					// 	m.Directory,
					// 	p,
					// ),
				)
				if err != nil {
					if os.IsNotExist(err) {
						data, err = templates.ReadFile(
							path.Join("/html/", p+".html"),
						)
						if err == nil {
							return data, nil
						}
					}
					return nil, fmt.Errorf("unable to load template %s: %w", p, err)
				}
				return data, nil
			}
		default:
			return nil, fmt.Errorf("unsupported template variable type: %T", p)
		}
	}

	data, err = templates.ReadFile("html/default.html")
	if err != nil {
		return nil, fmt.Errorf("unable to load the default template: %w", err)
	}
	return data, err
}
