package template

import "html/template"

func WithNewRootTemplate(target, replacement *template.Template) *template.Template {
	var err error
	target, err = target.AddParseTree("", replacement.Tree)
	if err != nil {
		panic(err)
	}
	return target
}
