package mdsend

import (
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"time"

	"github.com/dkotik/mdsend/header"
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

// Message is an intent to delivery a copy of a letter to a particular recipient.
type Message struct {
	ID            string
	LetterID      string
	SeedKey       string
	Headers       []header.Header
	From          mail.Address
	To            mail.Address
	Subject       string
	Text          string
	HTML          string
	ScheduleAfter time.Time
	ScheduledAt   time.Time
	SentAt        time.Time
}

func (m Message) Validate() (err error) {
	// if m.ID == "" {
	// 	return errors.New("empty ID")
	// }
	if m.LetterID == "" {
		return errors.New("empty letter ID")
	}
	for _, h := range m.Headers {
		if _, err = header.New(h.Name, h.Value); err != nil {
			return err
		}

	}
	if m.From.Address == "" {
		return errors.New("empty from address")
	}
	if m.To.Address == "" {
		return errors.New("empty recipient address")
	}
	if m.Subject == "" {
		return errors.New("empty subject")
	}
	if m.Text == "" {
		return errors.New("empty plain text content")
	}
	if m.HTML == "" {
		return errors.New("empty HTML content")
	}
	return nil
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
	if m.SeedKey != b.SeedKey {
		return FieldComparisonMismatchError{
			FieldName:     "SeedKey",
			ExpectedValue: m.SeedKey,
			ActualValue:   b.SeedKey,
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
