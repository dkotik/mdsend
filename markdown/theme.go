package markdown

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type Color struct {
	Action     string
	Heading    string
	Text       string
	Link       string
	BlockQuote string
	Border     string // thematic break, tables, code block, blockquote margin
	Table      string // table row background
	Shadow     string // blockquote, table alt row background, inline code
}

type Theme struct {
	Color Color

	FontFamily string
	FontSize   uint8
}

var (
	DefaultLightTheme = Theme{
		Color: Color{
			Action:     "#3a86ff",
			Heading:    "#11267",
			Text:       "#11226a",
			Link:       "#335c9c",
			BlockQuote: "#e8ecfb",
			Border:     "#335c4c",
			Table:      "#e8ecfb",
			Shadow:     "#e8ecfb",
		},
		FontFamily: "Georgia",
		FontSize:   12,
	}
)

func (t Theme) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	err := ast.Walk(node, func(node ast.Node, enter bool) (ast.WalkStatus, error) {
		switch node.Kind() {
		case ast.KindAutoLink:
			fallthrough
		case ast.KindLink:
			if t.Color.Link != "" {
				ApplyStyle(node, "color:"+t.Color.Link+";")
			}
		case ast.KindBlockquote:
			style := "padding: 6px 14px 6px 12px;"
			if t.Color.BlockQuote != "" {
				style += "color:" + t.Color.BlockQuote + ";"
			}
			if t.Color.Border != "" {
				style += "border:none;border-left:2px solid " + t.Color.Border + ";"
			}
			if t.Color.Shadow != "" {
				style += "background-color:" + t.Color.Shadow + ";"
			}
			if t.FontFamily != "" {
				style += "font-family:" + t.FontFamily + ";"
			}
			ApplyStyle(node, style)
		case ast.KindCodeBlock:
			style := "font-family:monospace;padding: 6px 14px 6px 12px;"
			if t.Color.Shadow != "" {
				style += "background-color:" + t.Color.Shadow + ";"
			}
			if t.FontSize != 0 {
				style += fmt.Sprintf("font-size: %dpx;", t.FontSize)
			}
			ApplyStyle(node, style)
		case ast.KindCodeSpan:
			style := "font-family:monospace;"
			if t.Color.Shadow != "" {
				style += "background-color:" + t.Color.Shadow + ";"
			}
			ApplyStyle(node, style)
		case ast.KindHeading:
			node := node.(*ast.Heading)
			style := ""
			if t.Color.Heading != "" {
				style += "color:" + t.Color.Heading + ";"
			}
			if t.FontFamily != "" {
				style += "font-family:" + t.FontFamily + ";"
			}
			if t.FontSize != 0 {
				switch node.Level {
				case 1:
					style += fmt.Sprintf("font-size: %dpx;", t.FontSize+12)
				case 2:
					style += fmt.Sprintf("font-size: %dpx;", t.FontSize+10)
				case 3:
					style += fmt.Sprintf("font-size: %dpx;", t.FontSize+8)
				case 4:
					style += fmt.Sprintf("font-size: %dpx;", t.FontSize+6)
				case 5:
					style += fmt.Sprintf("font-size: %dpx;", t.FontSize+4)
				case 6:
					style += fmt.Sprintf("font-size: %dpx;", t.FontSize+2)
				}
			}
			ApplyStyle(node, style)
		case ast.KindImage:
		case ast.KindList:
		case ast.KindListItem:
		case ast.KindParagraph:
			style := ""
			if t.FontSize != 0 {
				style += fmt.Sprintf("font-size:%dpx;", t.FontSize)
			}
			if t.FontFamily != "" {
				style += "font-family:" + t.FontFamily + ";"
			}
			if t.Color.Text != "" {
				style += "color:" + t.Color.Text + ";"
			}
			ApplyStyle(node, style)
		case ast.KindThematicBreak:
			style := "padding:10px 25px 10px 25px;"
			if t.Color.Border != "" {
				style += "border:none;border-top:2px solid " + t.Color.Border + ";"
			}
			ApplyStyle(node, style)
		case KindActionButton:
			// 	options.ActionContainerStyle = "border-radius: 5px; background-color:#3a86ff;"
			// 	options.ActionStyle = "font-size: 18px; color: #ffffff; font-weight: bold; text-decoration: none;border-radius: 5px; padding: 12px 18px; border: 1px solid #3a86ff; display: inline-block;"
			style := ""
			if t.Color.Action != "" {
				style += "border-radius:5px;background-color:" + t.Color.Action + ";width:80%;max-width:250px;"
			}
			ApplyStyle(node, style)
			style = fmt.Sprintf("font-size:%dpx;", min(18, t.FontSize+6))
			style += "font-weight:bold;text-decoration:none;border-radius:5px;padding:12px 18px;display:inline-block;"
			if t.Color.Action != "" {
				style += "color:#ffffff;border:1px solid " + t.Color.Action + ";"
			}
			if t.FontFamily != "" {
				style += "font-family:" + t.FontFamily + ";"
			}
			link := node.FirstChild().(*ast.Link)
			ApplyStyle(link, style)
			return ast.WalkSkipChildren, nil
			// case ast.KindDocument:
			// case ast.KindEmphasis:
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		panic(err)
	}
}

var styleAttributeName = []byte(`style`)

func ApplyStyle(n ast.Node, style string) {
	currentRaw, ok := n.Attribute(styleAttributeName)
	if ok {
		switch currentRaw := currentRaw.(type) {
		case nil:
		case string:
			n.SetAttribute(styleAttributeName, currentRaw+style)
		default:
		}
	}
	n.SetAttribute(styleAttributeName, style)
}
