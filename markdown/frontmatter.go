package markdown

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

const (
	IdempotentIdentifierKey = "id"
	FromKey                 = "from"
	ToKey                   = "to"
	CarbonCopyKey           = "cc"
	BlindCopyKey            = "bcc"
	SubjectKey              = "subject"
	AttachmentsKey          = "attachments"
	TemplatesKey            = "templates"
	NameKey                 = "name"
	EmailKey                = "email"

	lineBreak      rune = '\n'
	carriageReturn rune = '\r'
	delimeterYAML  rune = '-'
	delimeterTOML  rune = '+'
)

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
	if delimeter != delimeterYAML && delimeter != delimeterTOML { // no frontmatter
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
