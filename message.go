package mdsend

import (
	"context"
	"fmt"
	"io"
	"net/mail"
	"net/textproto"
	"sort"
	"strings"
	"time"
)

type Message interface {
	To() []string
	CC() []string
	BCC() []string
	Subject() string
	MIMEBody() io.ReadCloser
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

type Dispatch struct {
	ID       string
	LetterID string
	Headers  []Header
	From     mail.Address
	To       mail.Address
	Subject  string
	Text     string
	HTML     string
	SentAt   time.Time
}

type Sender interface {
	Send(context.Context, Dispatch) error
}
