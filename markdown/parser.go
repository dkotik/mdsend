package markdown

import (
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

func NewParser(theme Theme) parser.Parser {
	return parser.NewParser(
		parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithBlockParsers(
			util.Prioritized(extension.NewDefinitionDescriptionParser(), 700),
		),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithInlineParsers(
			util.Prioritized(extension.NewLinkifyParser(), 500),
			util.Prioritized(extension.NewFootnoteParser(), 600),
			util.Prioritized(extension.NewTaskCheckBoxParser(), 700),
			util.Prioritized(extension.NewTypographerParser(), 9999),
		),
		parser.WithParagraphTransformers(
			parser.DefaultParagraphTransformers()...,
		),
		parser.WithParagraphTransformers(
			util.Prioritized(extension.NewTableParagraphTransformer(), 500),
		),
		parser.WithASTTransformers(
			util.Prioritized(&ActionButtonInjector{}, 100),
			util.Prioritized(extension.NewTableASTTransformer(), 200),
			util.Prioritized(theme, 1000),
		),
	)
}
