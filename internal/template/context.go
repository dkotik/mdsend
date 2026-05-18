package template

import "html/template"

type Context struct {
	Frontmatter map[string]any
	Recipient   map[string]any
	Content     template.HTML
}
