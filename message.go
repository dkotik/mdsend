package mdsend

import (
	"fmt"
	"net/mail"
	"net/textproto"
	"sort"
	"strings"
	"time"
)

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
