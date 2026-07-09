package template

import (
	"html/template"

	"github.com/dkotik/mdsend"
)

type Context struct {
	Frontmatter map[string]any
	Recipient   map[string]any
	Content     template.HTML
	Schedule    mdsend.Schedule
}
