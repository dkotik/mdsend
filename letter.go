package mdsend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/mail"
	"path"
	"strings"
	"time"

	"github.com/dkotik/mdsend/internal/media"
	"github.com/dkotik/mdsend/markdown"
)

type LetterError uint

const (
	ErrNoSubject LetterError = iota + 1
	// ErrNoQueue
	ErrNoFromAddress
	ErrNoContent
	ErrLetterNotFound
)

func (l LetterError) Error() string {
	switch l {
	case ErrNoSubject:
		return "there is no subject in frontmatter"
	case ErrNoFromAddress:
		return "there is no from address in frontmatter"
	case ErrNoContent:
		return "there is no content"
	case ErrLetterNotFound:
		return "letter not found"
	default:
		return "unknown error"
	}
}

type Letter struct {
	ID          string
	Frontmatter map[string]any
	Content     string
	CreatedAt   time.Time
	SentAt      time.Time
	// MessageCount int
}

func NewLetterFromFile(
	ctx context.Context,
	fs fs.FS,
	p string,
) (letter Letter, err error) {
	file, err := fs.Open(p)
	if err != nil {
		return letter, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return letter, errors.Join(err, file.Close())
	}
	if err = file.Close(); err != nil {
		return letter, err
	}
	letter, err = NewLetter(data)
	if err != nil {
		return letter, err
	}
	letter, err = extend(ctx, letter, path.Dir(p), media.NewCyclicalImportPreventingFileSystem(fs))
	return letter, err
}

func NewLetter(b []byte) (letter Letter, err error) {
	frontmatterRaw, body, delimeter, err := splitFrontmatterFromContent(b)
	if err != nil {
		return letter, err
	}
	frontmatter, err := parseFrontmatter(frontmatterRaw, delimeter)
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

func (l Letter) GetQueue() string {
	switch queue := l.Frontmatter[FieldNameQueue].(type) {
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

func (l Letter) GetSubject() (string, error) {
	switch subject := l.Frontmatter[FieldNameSubject].(type) {
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

// GetFrom returns the [FieldNameFrom] address from the frontmatter.
// There must be only one valid address. Mutliple from addresses
// can disrupt delivery.
func (l Letter) GetFrom() (mail.Address, error) {
	switch from := l.Frontmatter[FieldNameFrom].(type) {
	case map[string]any:
		return newAddressFromMap(from)
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

func (l Letter) AssertEqualityTo(b Letter) error {
	if l.ID != b.ID {
		return FieldComparisonMismatchError{
			FieldName:     "ID",
			ExpectedValue: l.ID,
			ActualValue:   b.ID,
		}
	}
	if !maps.Equal(l.Frontmatter, b.Frontmatter) {
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
