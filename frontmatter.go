package mdsend

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
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

const (
	FieldNameID                        = "id"
	FieldNameExtends                   = "extends"
	FieldNameDatabase                  = "queue"
	FieldNameSubject                   = "subject"
	FieldNameFrom                      = "from"
	FieldNameReplyTo                   = "reply_to"
	FieldNameTo                        = "to"
	FieldNameCarbonCopy                = "cc"
	FieldNameBlindCarbonCopy           = "bcc"
	FieldNameName                      = "name"
	FieldNameEmail                     = "email"
	FieldNameAttachments               = "attachments"
	FieldNameAttachmentName            = "name"
	FieldNameAttachmentLocation        = "location"
	FieldNameTemplates                 = "templates"
	FieldNameHeaders                   = "headers"
	FieldNameLanguage                  = "language"
	FieldNameMediaContraints           = "media_constraints"
	FieldNameMediaConstraintsQuality   = "quality"
	FieldNameMediaConstrainsResolution = "resolution"
	FieldNameMediaConstraintsWidth     = "width"
	FieldNameMediaConstrainsHeight     = "height"
	FieldNameSchedule                  = "schedule"
	FieldNameScheduleAfter             = "after"
	FieldNameScheduleDelay             = "delay"
	FieldNameScheduleStep              = "step"
	FieldNameScheduleExpire            = "expire"
	FieldNameScheduleFluctuate         = "fluctuate"
	FieldNameListID                    = "list_id"
	FieldNameUnsubscribe               = "unsubscribe"
	FieldNameUnsubscribeEmail          = "unsubscribe_email"
	FieldNameUnsubscribeURL            = "unsubscribe_url"
)

func parseFrontmatter(source []byte, delimeter rune) (frontmatter map[string]any, err error) {
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

func splitFrontmatterFromContent(source []byte) (frontmatter, content []byte, delimeter rune, err error) {
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

func getIntFromMap(m map[string]interface{}, key string, defaultValue int) (int, error) {
	switch v := m[key].(type) {
	case nil:
		return defaultValue, nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case int16:
		return int(v), nil
	// case int8:
	// 	return int(v), nil
	// case uint8:
	// 	return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("invalid type for %s: %T", key, v)
	}
}

var rePercent = regexp.MustCompile(`^\s*(\d+)\s*\%\s*$`)

func getPercentageFromMap(m map[string]interface{}, key string, defaultValue int) (int, error) {
	switch v := m[key].(type) {
	case nil:
		return defaultValue, nil
	case string:
		m := rePercent.FindStringSubmatch(v)
		if m == nil {
			return defaultValue, fmt.Errorf("invalid percentage for %s: %s", key, v)
		}
		percent, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, fmt.Errorf("invalid percentage for %s: %s", key, v)
		}
		if percent > 100 {
			return 0, fmt.Errorf("invalid percentage for %s: %s", key, v)
		}
		return percent, nil
	default:
		return 0, fmt.Errorf("invalid type for %s: %T", key, v)
	}
}
