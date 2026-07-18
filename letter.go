package mdsend

import (
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/dkotik/mdsend/address"
	"github.com/dkotik/mdsend/header"
	"github.com/dkotik/mdsend/internal/media"
	"github.com/dkotik/mdsend/markdown"
	"golang.org/x/text/language"
)

var _ slog.LogValuer = (*Letter)(nil)

type LetterError uint

const (
	ErrNoSubject LetterError = iota + 1
	// ErrNoQueue
	ErrDuplicateLetter
	ErrNoFromAddress
	ErrNoContent
	ErrNotFound
	ErrFieldNotFound
)

func (l LetterError) Error() string {
	switch l {
	case ErrNoSubject:
		return "there is no subject in frontmatter"
	case ErrDuplicateLetter:
		return "duplicate letter"
	case ErrNoFromAddress:
		return "there is no from address in frontmatter"
	case ErrNoContent:
		return "there is no content"
	case ErrNotFound:
		return "letter not found"
	default:
		return "unknown error"
	}
}

type Letter struct {
	ID string
	// Frontmatter orderedmap.OrderedMap[string, any]
	Frontmatter map[string]any
	Templates   []Attachment
	Content     string
	CreatedAt   time.Time
	SentAt      time.Time
}

func NewLetter(b []byte) (letter Letter, err error) {
	letter, err = newLetter(b)
	if err != nil {
		return letter, err
	}

	if _, err = newSubject(letter.Frontmatter[FieldNameSubject]); err != nil {
		if errors.Is(err, ErrNoSubject) {
			// pull the subject from the first heading text
			letter.Frontmatter[FieldNameSubject] = markdown.GetFirstHeadingText([]byte(letter.Content))
			if letter.Frontmatter[FieldNameSubject] == "" {
				return letter, err
			}
		} else {
			return letter, err
		}
	}

	return letter, nil
}

// newLetter is a lighter version of NewLetter used for loading
// Markdown documents meant for extension.
func newLetter(b []byte) (letter Letter, err error) {
	frontmatterRaw, body, delimeter, err := markdown.SplitFrontmatterFromContent(b)
	if err != nil {
		return letter, err
	}
	frontmatter, err := markdown.ParseFrontmatter(frontmatterRaw, delimeter)
	if err != nil {
		return letter, err
	}
	if links := markdown.CollectLinks(body); len(links) > 0 {
		maps := make([]any, len(links))
		for i, link := range links {
			maps[i] = map[string]any{
				"name":     link.Name,
				"location": link.Destination,
			}
		}
		switch attachments := frontmatter[FieldNameAttachments].(type) {
		case nil:
			frontmatter[FieldNameAttachments] = maps
		case []any:
			frontmatter[FieldNameAttachments] = append(attachments, maps...)
		default:
			frontmatter[FieldNameAttachments] = append([]any{attachments}, maps...)
		}
	}

	return Letter{
		Frontmatter: frontmatter,
		Content:     string(body),
	}, nil
}

func (l Letter) GetDatabase() string {
	switch queue := l.Frontmatter[FieldNameDatabase].(type) {
	case string:
		queue = strings.TrimSpace(queue)
		if len(queue) == 0 {
			return ""
		}
		return queue
	default:
		return ""
	}
}

func (l Letter) GetSeed() (string, error) {
	switch seed := l.Frontmatter[FieldNameSeed].(type) {
	case nil:
		return "", nil
	case string:
		return strings.TrimSpace(seed), nil
	default:
		return strings.TrimSpace(fmt.Sprintf("%+v", seed)), nil
	}
}

func newSubject(v any) (string, error) {
	switch subject := v.(type) {
	case int, uint, int64, uint64, uint16, int16, float32, float64:
		numeric := fmt.Sprintf("%v", subject)
		if len(numeric) == 0 {
			return "", ErrNoSubject
		}
		return numeric, nil
	case string:
		subject = strings.TrimSpace(subject)
		if len(subject) == 0 {
			return "", ErrNoSubject
		}
		return subject, nil
	default:
		return "", ErrNoSubject
	}
}

func (l Letter) GetSubject() (string, error) {
	return newSubject(l.Frontmatter[FieldNameSubject])
}

// GetFrom returns the [address.FieldFrom] address from the frontmatter.
// There must be only one valid address. Mutliple from addresses
// can disrupt delivery.
func (l Letter) GetFrom() (mail.Address, error) {
	switch from := l.Frontmatter[address.FieldFrom].(type) {
	case map[string]any:
		return address.New(from)
	case string:
		if strings.TrimSpace(from) == "" {
			return mail.Address{}, ErrNoFromAddress
		}
		address, err := mail.ParseAddress(from)
		if err != nil {
			return mail.Address{}, err
		}
		return *address, nil
	default:
		return mail.Address{}, ErrNoFromAddress
	}
}

func getTemplates(frontmatter map[string]any, prefix string) (templates []string, err error) {
	switch t := frontmatter[FieldNameTemplates].(type) {
	case nil:
	case string:
		t = strings.TrimSpace(t)
		if len(t) > 0 {
			t = filepath.Join(prefix, t)
			templates = append(templates, t)
		}
	case []any:
		for _, v := range t {
			switch s := v.(type) {
			case string:
				s = strings.TrimSpace(s)
				if len(s) > 0 {
					s = filepath.Join(prefix, s)
					templates = append(templates, s)
				}
			default:
				continue
			}
		}
	default:
		return templates, fmt.Errorf(
			"invalid templates type: %+v (%T)",
			t, t,
		)
	}
	return templates, nil
}

func (l Letter) GetHeaders() (headers []header.Header, err error) {
	switch h := l.Frontmatter[FieldNameHeaders].(type) {
	case nil:
		return headers, nil
	case map[string]any:
		for name, value := range h {
			switch value := value.(type) {
			case int64, uint64, int32, uint32, int16, uint16, int8, uint8:
				header, err := header.New(name, fmt.Sprintf("%d", value))
				if err != nil {
					return headers, err
				}
				headers = append(headers, header)
			case float64, float32:
				header, err := header.New(name, fmt.Sprintf("%v", value))
				if err != nil {
					return headers, err
				}
				headers = append(headers, header)
			case string:
				header, err := header.New(name, fmt.Sprintf("%v", value))
				if err != nil {
					return headers, err
				}
				headers = append(headers, header)
			default:
				return headers, fmt.Errorf(
					"header %q has invalid value type: %+v (%T)",
					name, value, value,
				)
			}
		}
		return headers, nil
	default:
		return headers, fmt.Errorf("letter headers are of wrong type: %+v (%T)", h, h)
	}
}

func (l Letter) GetLanguage() (lang language.Tag, err error) {
	switch languageTag := l.Frontmatter[FieldNameLanguage].(type) {
	case nil:
	case string:
		// languageTag = strings.TrimSpace(languageTag)
		lang, err = language.Parse(languageTag)
		if err != nil {
			return lang, err
		}
		return lang, nil
	default:
		return lang, fmt.Errorf("invalid language type: %T", languageTag)
	}
	return lang, nil
}

func (l Letter) GetMediaConstraints() (m media.Constraints, err error) {
	switch media := l.Frontmatter[FieldNameMediaContraints].(type) {
	case nil:
		return m, nil
	case map[string]any:
		m.Quality, err = getPercentageFromMap(media, FieldNameMediaConstraintsQuality, 80)
		if err != nil {
			return m, err
		}
		resolution, err := getIntFromMap(media, FieldNameMediaConstrainsResolution, 1080)
		if err != nil {
			return m, err
		}
		if resolution < 160 {
			return m, errors.New("resolution must be at least 160")
		}
		if resolution > 7680 {
			return m, fmt.Errorf("resolution must be at most 7680")
		}
		m = m.WithResolution(resolution)
		m.Width, err = getIntFromMap(media, FieldNameMediaConstraintsWidth, m.Width)
		if err != nil {
			return m, err
		}
		m.Height, err = getIntFromMap(media, FieldNameMediaConstrainsHeight, m.Height)
		if err != nil {
			return m, err
		}
		return m, nil
	default:
		return m, fmt.Errorf("invalid media constraints %T: %v", media, media)
	}
}

func (l Letter) AssertEqualityTo(b Letter) error {
	if l.ID != b.ID {
		return FieldComparisonMismatchError{
			FieldName:     "ID",
			ExpectedValue: l.ID,
			ActualValue:   b.ID,
		}
	}
	if !reflect.DeepEqual(l.Frontmatter, b.Frontmatter) {
		return FieldComparisonMismatchError{
			FieldName:     "Frontmatter",
			ExpectedValue: fmt.Sprintf("%+v", l.Frontmatter),
			ActualValue:   fmt.Sprintf("%+v", b.Frontmatter),
		}
	}
	if l.Content != b.Content {
		return FieldComparisonMismatchError{
			FieldName:     "Content",
			ExpectedValue: l.Content,
			ActualValue:   b.Content,
		}
	}
	if len(l.Templates) != len(b.Templates) {
		return FieldComparisonMismatchError{
			FieldName:     "Templates",
			ExpectedValue: fmt.Sprintf("%d templates", len(l.Templates)),
			ActualValue:   fmt.Sprintf("%d templates", len(b.Templates)),
		}
	}
	for i := range l.Templates {
		if err := l.Templates[i].AssertEqualityTo(b.Templates[i]); err != nil {
			return FieldComparisonMismatchError{
				FieldName:     fmt.Sprintf("Templates[%d]", i),
				ExpectedValue: fmt.Sprintf("%+v", l.Templates[i]),
				ActualValue:   fmt.Sprintf("%+v", b.Templates[i]),
			}
		}
	}
	if !l.CreatedAt.Truncate(time.Second).Equal(b.CreatedAt.Truncate(time.Second)) {
		return FieldComparisonMismatchError{
			FieldName:     "CreatedAt",
			ExpectedValue: l.CreatedAt.Format(time.RFC3339),
			ActualValue:   b.CreatedAt.Format(time.RFC3339),
		}
	}
	if !l.SentAt.Truncate(time.Second).Equal(b.SentAt.Truncate(time.Second)) {
		return FieldComparisonMismatchError{
			FieldName:     "SentAt",
			ExpectedValue: l.SentAt.Format(time.RFC3339),
			ActualValue:   b.SentAt.Format(time.RFC3339),
		}
	}
	return nil
}

func (l Letter) IsEqualTo(b Letter) bool {
	return l.AssertEqualityTo(b) == nil
}

func (l Letter) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("ID", l.ID),
		slog.Any("subject", l.Frontmatter[FieldNameSubject]),
	)
}
