package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"

	"cuelang.org/go/cue/cuecontext"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

const (
	lineBreak      rune = '\n'
	carriageReturn rune = '\r'

	FontmatterDelimeterYAML rune = '-'
	FrontmaterDelimeterTOML rune = '+'
	FrontmaterDelimeterCue  rune = ':'
)

func ParseFrontmatter(source []byte, delimeter rune) (frontmatter map[string]any, err error) {
	switch delimeter {
	case FontmatterDelimeterYAML:
		if err = yaml.NewDecoder(bytes.NewReader(source)).Decode(&frontmatter); err != nil {
			return nil, fmt.Errorf("invalid YAML front-matter: %w", err)
		}
	case FrontmaterDelimeterTOML:
		if err = toml.NewDecoder(bytes.NewReader(source)).Decode(&frontmatter); err != nil {
			return nil, fmt.Errorf("invalid TOML front-matter: %w", err)
		}
	case FrontmaterDelimeterCue:
		ctx := cuecontext.New()
		if err = ctx.CompileBytes(source).Decode(&frontmatter); err != nil {
			return nil, fmt.Errorf("invalid CUE front-matter: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported front-matter delimeter: %s", string(delimeter))
	}
	return frontmatter, nil
}

func Cut(source []byte) (frontmatter, content []byte, delimeter rune, err error) {
	var (
		r      = bufio.NewReader(bytes.NewReader(source))
		cursor = 0
		count  = 0
	)

	for { // drain white space
		delimeter, count, err = r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return nil, source, delimeter, nil
			}
			return nil, source, delimeter, err
		}
		if !unicode.IsSpace(delimeter) {
			source = source[cursor:] // cut off white space
			cursor = count           // reset cursor
			break
		}
		cursor = cursor + count
	}
	if delimeter != FontmatterDelimeterYAML && delimeter != FrontmaterDelimeterTOML { // no frontmatter
		return nil, source, delimeter, nil
	}

	var (
		next                  rune
		openingDelimeterCount int
	)
	for { // drain opening front-matter marker
		next, count, err = r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return nil, source, delimeter, nil
			}
			return nil, source, delimeter, err
		}
		if next != delimeter {
			if cursor < 3 { // not enough delimeters
				return nil, source, delimeter, nil
			}
			openingDelimeterCount = cursor
			break
		}
		cursor = cursor + count
	}

	newLine := false
	for { // drain trailing white space after the marker
		cursor = cursor + count
		if !unicode.IsSpace(next) {
			break
		}
		if next == lineBreak {
			newLine = true
		}
		next, count, err = r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return nil, source, delimeter, nil
			}
			return nil, source, delimeter, err
		}
	}

	if !newLine {
		return nil, source, delimeter, nil
	}
	frontmatter = source[cursor-count:]
	content = frontmatter
	cursor = count
	closingDelimeterCount := 0
	extraWhiteSpace := 0
	if next == delimeter {
		closingDelimeterCount = 1
	} else {
		newLine = false
	}

	for { // find closing
		next, count, err = r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return nil, source, delimeter, nil
			}
			return nil, source, delimeter, err
		}
		cursor = cursor + count

		if newLine {
			if next == delimeter {
				closingDelimeterCount++
				if openingDelimeterCount == closingDelimeterCount {
					frontmatter = frontmatter[:cursor-openingDelimeterCount-extraWhiteSpace]
					break // closing candidate found
				}
				continue
			}
			if next == carriageReturn && closingDelimeterCount == 0 {
				continue // ignore carriage returns after new lines
			}
			newLine = false
		}
		if unicode.IsSpace(next) {
			if next == lineBreak {
				newLine = true
			}
			extraWhiteSpace += count
		} else {
			extraWhiteSpace = 0
		}
		closingDelimeterCount = 0
	}

	newLine = false
	for { // drain trailing white space after the closing marker
		next, count, err = r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return nil, source, delimeter, nil
			}
			return nil, source, delimeter, err
		}
		cursor = cursor + count
		if !unicode.IsSpace(next) {
			break
		}
		if next == lineBreak {
			newLine = true
		}
	}
	if !newLine {
		return nil, source, delimeter, nil // closing line end not found
	}
	return frontmatter, content[cursor-count:], delimeter, nil
}
