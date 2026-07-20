package markdown

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	reValidHexColor   = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
	reValidFontFamily = regexp.MustCompile(`^(?i)[a-z0-9\- ]{4,}$`)
)

type Palette struct {
	Action     string
	Heading    string
	Text       string
	Link       string
	BlockQuote string
	Border     string // thematic break, tables, code block, blockquote margin
	Table      string // table row background
	Shadow     string // blockquote, table alt row background, inline code
}

func NewColor(s string) (string, error) {
	if !reValidHexColor.MatchString(s) {
		namedValue, ok := validNamedColors[s]
		if !ok {
			return "", fmt.Errorf("invalid hex color: %s", s)
		}
		return namedValue, nil
	}
	return s, nil
}

func NewColorFromAny(v any) (string, error) {
	switch v := v.(type) {
	case nil:
		return "", nil
	case string:
		return NewColor(v)
	default:
		return "", fmt.Errorf("invalid color type: %v (%T)", v, v)
	}
}

func NewPaletteFromMap(m map[string]any) (p Palette, err error) {
	for k, v := range m {
		switch k {
		case "action":
			p.Action, err = NewColorFromAny(v)
			if err != nil {
				return p, err
			}
		case "heading":
			p.Heading, err = NewColorFromAny(v)
			if err != nil {
				return p, err
			}
		case "text":
			p.Text, err = NewColorFromAny(v)
			if err != nil {
				return p, err
			}
		case "link":
			p.Link, err = NewColorFromAny(v)
			if err != nil {
				return p, err
			}
		case "blockquote":
			p.BlockQuote, err = NewColorFromAny(v)
			if err != nil {
				return p, err
			}
		case "border":
			p.Border, err = NewColorFromAny(v)
			if err != nil {
				return p, err
			}
		case "table":
			p.Table, err = NewColorFromAny(v)
			if err != nil {
				return p, err
			}
		case "shadow":
			p.Shadow, err = NewColorFromAny(v)
			if err != nil {
				return p, err
			}
		}
	}
	return p, nil
}

type Theme struct {
	Color Palette

	FontFamily string
	FontSize   uint8
}

func NewThemeFromMap(m map[string]any) (t Theme, err error) {
	for k, v := range m {
		switch k {
		case "color":
			m, ok := v.(map[string]any)
			if ok {
				t.Color, err = NewPaletteFromMap(m)
				if err != nil {
					return t, err
				}
			}
		case "font_family":
			ff, ok := v.(string)
			if !ok {
				return t, fmt.Errorf("invalid font_family: %v (%T)", v, v)
			}
			if !IsValidFontFamily(ff) {
				return t, fmt.Errorf("invalid font_family format: %s", ff)
			}
			t.FontFamily = strings.ReplaceAll(ff, "\"", "'")
		case "font_size":
			fs, _ := v.(string)
			fs = strings.TrimSpace(strings.ToLower(fs))
			if !strings.HasSuffix(fs, "px") {
				return t, fmt.Errorf("invalid font size: electronic mail font size must be specified in pixels (px): %s", fs)
			}
			fsUint, err := strconv.ParseUint(strings.TrimSpace(fs[:len(fs)-2]), 10, 8)
			if err != nil {
				return t, fmt.Errorf("invalid font size: %w", err)
			}
			if fsUint > 72 {
				return t, fmt.Errorf("invalid font size: %d is too large (max 72)", fsUint)
			}
			if fsUint < 8 {
				return t, fmt.Errorf("invalid font size: %d is too small (min 8)", fsUint)
			}
			t.FontSize = uint8(fsUint)
		}
	}
	return t, nil
}

var (
	DefaultLightTheme = Theme{
		Color: Palette{
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

var validNamedColors = map[string]string{
	"black":                "#000000",
	"silver":               "#c0c0c0",
	"gray":                 "#808080",
	"white":                "#ffffff",
	"maroon":               "#800000",
	"red":                  "#ff0000",
	"purple":               "#800080",
	"fuchsia":              "#ff00ff",
	"green":                "#008000",
	"lime":                 "#00ff00",
	"olive":                "#808000",
	"yellow":               "#ffff00",
	"navy":                 "#000080",
	"blue":                 "#0000ff",
	"teal":                 "#008080",
	"aqua":                 "#00ffff",
	"aliceblue":            "#f0f8ff",
	"antiquewhite":         "#faebd7",
	"aquamarine":           "#7fffd4",
	"azure":                "#f0ffff",
	"beige":                "#f5f5dc",
	"bisque":               "#ffe4c4",
	"blanchedalmond":       "#ffebcd",
	"blueviolet":           "#8a2be2",
	"brown":                "#a52a2a",
	"burlywood":            "#deb887",
	"cadetblue":            "#5f9ea0",
	"chartreuse":           "#7fff00",
	"chocolate":            "#d2691e",
	"coral":                "#ff7f50",
	"cornflowerblue":       "#6495ed",
	"cornsilk":             "#fff8dc",
	"crimson":              "#dc143c",
	"cyan":                 "#00ffff",
	"darkblue":             "#00008b",
	"darkcyan":             "#008b8b",
	"darkgoldenrod":        "#b8860b",
	"darkgray":             "#a9a9a9",
	"darkgreen":            "#006400",
	"darkgrey":             "#a9a9a9",
	"darkkhaki":            "#bdb76b",
	"darkmagenta":          "#8b008b",
	"darkolivegreen":       "#556b2f",
	"darkorange":           "#ff8c00",
	"darkorchid":           "#9932cc",
	"darkred":              "#8b0000",
	"darksalmon":           "#e9967a",
	"darkseagreen":         "#8fbc8f",
	"darkslateblue":        "#483d8b",
	"darkslategray":        "#2f4f4f",
	"darkslategrey":        "#2f4f4f",
	"darkturquoise":        "#00ced1",
	"darkviolet":           "#9400d3",
	"deeppink":             "#ff1493",
	"deepskyblue":          "#00bfff",
	"dimgray":              "#696969",
	"dimgrey":              "#696969",
	"dodgerblue":           "#1e90ff",
	"firebrick":            "#b22222",
	"floralwhite":          "#fffaf0",
	"forestgreen":          "#228b22",
	"gainsboro":            "#dcdcdc",
	"ghostwhite":           "#f8f8ff",
	"gold":                 "#ffd700",
	"goldenrod":            "#daa520",
	"greenyellow":          "#adff2f",
	"grey":                 "#808080",
	"honeydew":             "#f0fff0",
	"hotpink":              "#ff69b4",
	"indianred":            "#cd5c5c",
	"indigo":               "#4b0082",
	"ivory":                "#fffff0",
	"khaki":                "#f0e68c",
	"lavender":             "#e6e6fa",
	"lavenderblush":        "#fff0f5",
	"lawngreen":            "#7cfc00",
	"lemonchiffon":         "#fffacd",
	"lightblue":            "#add8e6",
	"lightcoral":           "#f08080",
	"lightcyan":            "#e0ffff",
	"lightgoldenrodyellow": "#fafad2",
	"lightgray":            "#d3d3d3",
	"lightgreen":           "#90ee90",
	"lightgrey":            "#d3d3d3",
	"lightpink":            "#ffb6c1",
	"lightsalmon":          "#ffa07a",
	"lightseagreen":        "#20b2aa",
	"lightskyblue":         "#87cefa",
	"lightslategray":       "#778899",
	"lightslategrey":       "#778899",
	"lightsteelblue":       "#b0c4de",
	"lightyellow":          "#ffffe0",
	"limegreen":            "#32cd32",
	"linen":                "#faf0e6",
	"magenta":              "#ff00ff",
	"mediumaquamarine":     "#66cdaa",
	"mediumblue":           "#0000cd",
	"mediumorchid":         "#ba55d3",
	"mediumpurple":         "#9370db",
	"mediumseagreen":       "#3cb371",
	"mediumslateblue":      "#7b68ee",
	"mediumspringgreen":    "#00fa9a",
	"mediumturquoise":      "#48d1cc",
	"mediumvioletred":      "#c71585",
	"midnightblue":         "#191970",
	"mintcream":            "#f5fffa",
	"mistyrose":            "#ffe4e1",
	"moccasin":             "#ffe4b5",
	"navajowhite":          "#ffdead",
	"oldlace":              "#fdf5e6",
	"olivedrab":            "#6b8e23",
	"orange":               "#ffa500",
	"orangered":            "#ff4500",
	"orchid":               "#da70d6",
	"palegoldenrod":        "#eee8aa",
	"palegreen":            "#98fb98",
	"paleturquoise":        "#afeeee",
	"palevioletred":        "#db7093",
	"papayawhip":           "#ffefd5",
	"peachpuff":            "#ffdab9",
	"peru":                 "#cd853f",
	"pink":                 "#ffc0cb",
	"plum":                 "#dda0dd",
	"powderblue":           "#b0e0e6",
	"rebeccapurple":        "#663399",
	"rosybrown":            "#bc8f8f",
	"royalblue":            "#4169e1",
	"saddlebrown":          "#8b4513",
	"salmon":               "#fa8072",
	"sandybrown":           "#f4a460",
	"seagreen":             "#2e8b57",
	"seashell":             "#fff5ee",
	"sienna":               "#a0522d",
	"skyblue":              "#87ceeb",
	"slateblue":            "#6a5acd",
	"slategray":            "#708090",
	"slategrey":            "#708090",
	"snow":                 "#fffafa",
	"springgreen":          "#00ff7f",
	"steelblue":            "#4682b4",
	"tan":                  "#d2b48c",
	"thistle":              "#d8bfd8",
	"tomato":               "#ff6347",
	"transparent":          "transparent",
	"turquoise":            "#40e0d0",
	"violet":               "#ee82ee",
	"wheat":                "#f5deb3",
	"whitesmoke":           "#f5f5f5",
	"yellowgreen":          "#9acd32",
}

func IsValidFontFamily(ff string) bool {
	for _, f := range strings.Split(ff, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			return false
		}
		if f[0] == '\'' {
			if len(f) < 2 || f[len(f)-1] != '\'' {
				return false
			}
			f = f[1 : len(f)-1]
		}
		if f[0] == '"' {
			if len(f) < 2 || f[len(f)-1] != '"' {
				return false
			}
			f = f[1 : len(f)-1]
		}
		if !reValidFontFamily.MatchString(f) {
			return false
		}
	}
	return true
}
