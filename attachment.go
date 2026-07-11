package mdsend

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"log/slog"
	"path"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/cespare/xxhash/v2"
	"github.com/dkotik/mdsend/internal/media"
)

var _ slog.LogValuer = (*Attachment)(nil)

type AttachmentError uint8

const (
	ErrInvalidAttachment AttachmentError = iota
	ErrDuplicateAttachment
)

func (err AttachmentError) Error() string {
	switch err {
	case ErrInvalidAttachment:
		return "invalid attachment"
	case ErrDuplicateAttachment:
		return "duplicate attachment"
	default:
		return ""
	}
}

type AttachmentSource struct {
	Name     string
	Location string
}

type Attachment struct {
	LetterID string
	Name     string
	Source   string
	// Hash                     uint64 // for XXHash2
	Hash string
	// mimeEncodedBase64Content []byte

	// ContentID is the ID of the inline attachment to use in the message.
	// It must conform to RFC 2392 format, including the angle brackets:
	// <hash@domain.com>. Leave empty for standalone attachments that
	// are not referenced by other parts of the message.
	ContentID   string
	ContentType string
	Content     []byte
}

func NewAttachment(b []byte, constraints media.Constraints) (a Attachment, err error) {
	a.Content, a.ContentType, err = constraints.ApplyTo(b)
	if err != nil {
		return a, err
	}
	a.Hash = media.DeterministicHashStringOf(a.Content)
	return a, nil
}

func NewAttachmentFromFile(fs fs.FS, p string, constraints media.Constraints) (a Attachment, err error) {
	file, err := fs.Open(p)
	if err != nil {
		return a, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	b, err := io.ReadAll(file)
	if err != nil {
		return a, err
	}
	a, err = NewAttachment(b, constraints)
	if err != nil {
		return a, err
	}
	a.Name = path.Base(p)
	return a, err
}

func (a Attachment) Validate() (err error) {
	if a.LetterID == "" {
		return errors.New("empty letter ID")
	}
	if a.Name == "" {
		return errors.New("empty name")
	}
	if a.ContentType == "" {
		return errors.New("empty content type")
	}
	if len(a.Content) == 0 {
		return errors.New("empty content")
	}
	return nil
}

func (a Attachment) WithUpdatedHash() Attachment {
	hash := xxhash.New()
	var err error
	if _, err = io.Copy(
		hash,
		bytes.NewReader(a.Content),
	); err != nil {
		panic(err)
	}
	a.Hash = base58.Encode(hash.Sum(nil))
	// a.Hash = hash.Sum64()
	return a
}

func (a Attachment) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", a.Name),
		slog.String("content_type", a.ContentType),
		slog.String("hash", a.Hash),
	)
}

func (a Attachment) AssertEqualityTo(b Attachment) error {
	if a.LetterID != b.LetterID {
		return FieldComparisonMismatchError{
			FieldName:     "LetterID",
			ExpectedValue: a.LetterID,
			ActualValue:   b.LetterID,
		}
	}
	if a.Name != b.Name {
		return FieldComparisonMismatchError{
			FieldName:     "Name",
			ExpectedValue: a.Name,
			ActualValue:   b.Name,
		}
	}
	if a.ContentID != b.ContentID {
		return FieldComparisonMismatchError{
			FieldName:     "ContentID",
			ExpectedValue: a.ContentID,
			ActualValue:   b.ContentID,
		}
	}
	if a.ContentType != b.ContentType {
		return FieldComparisonMismatchError{
			FieldName:     "Name",
			ExpectedValue: a.ContentType,
			ActualValue:   b.ContentType,
		}
	}
	if !bytes.Equal(a.Content, b.Content) {
		return FieldComparisonMismatchError{
			FieldName:     "Name",
			ExpectedValue: a.Content,
			ActualValue:   b.Content,
		}
	}
	return nil
}

func (a Attachment) IsEqualTo(b Attachment) bool {
	return a.AssertEqualityTo(b) == nil
}

func newAttachmentSourceFromMap(fm map[string]any) (a AttachmentSource, _ error) {
	a.Location = strings.TrimSpace(fmt.Sprintf("%v", fm[FieldNameAttachmentLocation]))
	switch name := fm[FieldNameAttachmentName].(type) {
	case string:
		a.Name = strings.TrimSpace(name)
		if a.Name == "" {
			a.Name = path.Base(a.Location)
		}
	case nil:
		a.Name = path.Base(a.Location)
	default:
		a.Name = strings.TrimSpace(fmt.Sprintf("%v", name))
		if a.Name == "" {
			a.Name = path.Base(a.Location)
		}
	}
	return a, nil
}

func newAttachmentSourceFromAny(fm any) (AttachmentSource, error) {
	switch fm := fm.(type) {
	case map[string]any:
		return newAttachmentSourceFromMap(fm)
	case string:
		fm = strings.TrimSpace(fm)
		return AttachmentSource{Name: path.Base(fm), Location: fm}, nil
	default:
		return AttachmentSource{}, fmt.Errorf("invalid attachment source: %T %+v", fm, fm)
	}
}

func (l Letter) EachAttachmentSource() iter.Seq2[AttachmentSource, error] {
	return func(yield func(AttachmentSource, error) bool) {
		switch fm := l.Frontmatter[FieldNameAttachments].(type) {
		case []any:
			for _, a := range fm {
				if !yield(newAttachmentSourceFromAny(a)) {
					return
				}
			}
		case map[string]any:
			if !yield(newAttachmentSourceFromMap(fm)) {
				return
			}
		case string:
			yield(newAttachmentSourceFromAny(fm))
		case nil:
		default:
			yield(AttachmentSource{}, fmt.Errorf("invalid attachment source: %T %+v", fm, fm))
		}
	}
}

func EachAttachmentSourceRelativeWithPathPrefix(prefix string, each iter.Seq2[AttachmentSource, error]) iter.Seq2[AttachmentSource, error] {
	if each == nil {
		panic("attachment source list is nil")
	}
	return func(yield func(AttachmentSource, error) bool) {
		for a, err := range each {
			if err != nil {
				if !yield(AttachmentSource{}, err) {
					return
				}
			}
			if media.IsPathLocal(a.Location) {
				a.Location = path.Join(prefix, a.Location)
			}
			if !yield(a, nil) {
				return
			}
		}
	}
}
