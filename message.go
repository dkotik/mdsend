package mdsend

import (
	"fmt"
	"log/slog"
	"net/mail"
	"net/textproto"
	"sort"
	"strings"
	"time"
)

var _ slog.LogValuer = (*Message)(nil)

type MessageError uint8

const (
	ErrInvalidMessage MessageError = iota
	ErrDuplicateMessage
	ErrMessageNotFound
)

func (err MessageError) Error() string {
	switch err {
	case ErrInvalidMessage:
		return "invalid message"
	case ErrDuplicateMessage:
		return "duplicate message"
	case ErrMessageNotFound:
		return "message not found"
	default:
		return ""
	}
}

type Header struct {
	Name  string
	Value string
}

func MergeHeaders(ms ...map[string]any) (result []Header) {
	for _, m := range ms {
		for k, v := range m {
			k = textproto.CanonicalMIMEHeaderKey(k)
			switch v := v.(type) {
			case string:
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				result = append(result, Header{Name: k, Value: v})
			case []string:
				for _, s := range v {
					s = strings.TrimSpace(s)
					if s == "" {
						continue
					}
					result = append(result, Header{Name: k, Value: s})
				}
			case nil: // skip nil values
			default:
				result = append(result, Header{Name: k, Value: fmt.Sprintf("%v", v)})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return
}

// Message is an intent to delivery a copy of a letter to a particular recipient.
type Message struct {
	ID            string
	LetterID      string
	Headers       []Header
	From          mail.Address
	To            mail.Address
	Subject       string
	Text          string
	HTML          string
	ScheduleAfter time.Time
	ScheduledAt   time.Time
	SentAt        time.Time
}

func (m Message) AssertEqualityTo(b Message) error {
	if m.ID != b.ID {
		return FieldComparisonMismatchError{
			FieldName:     "ID",
			ExpectedValue: m.ID,
			ActualValue:   b.ID,
		}
	}
	if m.LetterID != b.LetterID {
		return FieldComparisonMismatchError{
			FieldName:     "LetterID",
			ExpectedValue: m.LetterID,
			ActualValue:   b.LetterID,
		}
	}
	if m.From.String() != b.From.String() {
		return FieldComparisonMismatchError{
			FieldName:     "From",
			ExpectedValue: m.From.String(),
			ActualValue:   b.From.String(),
		}
	}
	if m.To.String() != b.To.String() {
		return FieldComparisonMismatchError{
			FieldName:     "To",
			ExpectedValue: m.To.String(),
			ActualValue:   b.To.String(),
		}
	}
	if len(m.Headers) != len(b.Headers) {
		return FieldComparisonMismatchError{
			FieldName:     "Headers",
			ExpectedValue: fmt.Sprintf("%+v", m.Headers),
			ActualValue:   fmt.Sprintf("%+v", b.Headers),
		}
	}
	for i, header := range m.Headers {
		if header.Name != b.Headers[i].Name {
			return FieldComparisonMismatchError{
				FieldName:     "Headers[" + header.Name + "]",
				ExpectedValue: header.Name,
				ActualValue:   header.Name,
			}
		}
		if header.Value != b.Headers[i].Value {
			return FieldComparisonMismatchError{
				FieldName:     "Headers[" + header.Name + "]",
				ExpectedValue: header.Value,
				ActualValue:   header.Value,
			}
		}
		if m.Subject != b.Subject {
			return FieldComparisonMismatchError{
				FieldName:     "Subject",
				ExpectedValue: m.Subject,
				ActualValue:   b.Subject,
			}
		}
		if m.Text != b.Text {
			return FieldComparisonMismatchError{
				FieldName:     "Text",
				ExpectedValue: m.Text,
				ActualValue:   b.Text,
			}
		}
		if m.HTML != b.HTML {
			return FieldComparisonMismatchError{
				FieldName:     "HTML",
				ExpectedValue: m.HTML,
				ActualValue:   b.HTML,
			}
		}
		if !m.ScheduleAfter.Truncate(time.Second).Equal(b.ScheduleAfter.Truncate(time.Second)) {
			return FieldComparisonMismatchError{
				FieldName:     "SentAt",
				ExpectedValue: m.ScheduleAfter.Format(time.RFC3339),
				ActualValue:   b.ScheduleAfter.Format(time.RFC3339),
			}
		}
		if !m.ScheduledAt.Truncate(time.Second).Equal(b.SentAt.Truncate(time.Second)) {
			return FieldComparisonMismatchError{
				FieldName:     "SentAt",
				ExpectedValue: m.ScheduledAt.Format(time.RFC3339),
				ActualValue:   b.ScheduledAt.Format(time.RFC3339),
			}
		}
		if !m.SentAt.Truncate(time.Second).Equal(b.SentAt.Truncate(time.Second)) {
			return FieldComparisonMismatchError{
				FieldName:     "SentAt",
				ExpectedValue: m.SentAt.Format(time.RFC3339),
				ActualValue:   b.SentAt.Format(time.RFC3339),
			}
		}
	}
	return nil
}

func (m Message) IsEqualTo(b Message) bool {
	return m.AssertEqualityTo(b) == nil
}

func (m Message) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", m.ID),
		slog.String("letter_id", m.LetterID),
		slog.String("subject", m.Subject),
	)
}
