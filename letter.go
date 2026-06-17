package mdsend

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
)

type LetterError uint

const (
	ErrNoSubject LetterError = iota + 1
	ErrNoQueue
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

func (l Letter) GetQueue() (string, error) {
	switch queue := l.Frontmatter[FieldNameQueue].(type) {
	case string:
		queue = strings.TrimSpace(queue)
		if len(queue) == 0 {
			return "", ErrNoQueue
		}
		return queue, nil
	default:
		return "", ErrNoQueue
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

func (l Letter) GetSendAfter() (time.Time, error) {
	switch sendAfter := l.Frontmatter[FieldNameSendAfter].(type) {
	case nil:
		return time.Time{}, nil
	case string:
		t, err := time.Parse(time.RFC3339, sendAfter)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	default:
		return time.Time{}, fmt.Errorf("invalid send after format: %T", sendAfter)
	}
}

func (l Letter) Validate() (err error) {
	// if l.ID == "" {
	// 	return errors.New("letter has no ID")
	// }
	if strings.TrimSpace(l.Content) == "" {
		return ErrNoContent
	}
	if _, err = l.GetSubject(); err != nil {
		return err
	}
	if _, err = l.GetFrom(); err != nil {
		return err
	}
	if _, err = l.GetSchedule(); err != nil {
		return err
	}
	return nil
}
