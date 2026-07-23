package header

import (
	"errors"
	"fmt"
	"net/textproto"
	"net/url"
	"regexp"
	"strings"

	"github.com/dkotik/mdsend/address"
)

const (
	MIMEVersion             = "MIME-Version"
	ContentID               = "Content-ID"
	ContentType             = "Content-Type"
	ContentDescription      = "Content-Description"
	ContentDisposition      = "Content-Disposition"
	ContentTransferEncoding = "Content-Transfer-Encoding"
	From                    = "From"
	To                      = "To"
	Subject                 = "Subject"
	Date                    = "Date"
	ListID                  = "List-Id"
	ListOwner               = "List-Owner"
	ListPost                = "List-Post"
	ListHelp                = "List-Help"
	ListUnsubscribe         = "List-Unsubscribe"

	// If multiple URLs are present in List-Unsubscribe, the inbox provider will apply the List-Unsubscribe-Post command to the first HTTPS URI it finds. (RFC 8058)
	ListUnsubscribePost = "List-Unsubscribe-Post"
)

type Header struct {
	Name  string
	Value string
}

func New(name, value string) (h Header, err error) {
	h.Name = textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(name))
	if h.Name == "" {
		return h, ErrEmptyName
	}
	h.Value = strings.TrimSpace(value)
	if h.Value == "" {
		return h, fmt.Errorf("invalid value for header <%s>: %w", h.Name, ErrEmptyValue)
	}
	if strings.HasPrefix("List-", h.Name) {
		switch h.Name {
		case "List-ID":
			if err = ValidateListID(h.Value); err != nil {
				return h, err
			}
		case "List-Unsubscribe":
			if err = ValidateListUnsubscribe(h.Value); err != nil {
				return h, err
			}
		case "List-Unsubscribe-Post":
			if err = ValidateListUnsubscribePost(h.Value); err != nil {
				return h, err
			}
		}
	}
	return h, nil
}

func (h Header) String() string {
	return fmt.Sprintf("%s: %s", h.Name, h.Value)
}

var reValidListID = regexp.MustCompile(`(\w+\s+)*\<[^\<\>]+\>$`)

func ValidateListID(value string) (err error) {
	if !reValidListID.MatchString(value) {
		return fmt.Errorf("invalid List-Id header format: %s", value)
	}
	return nil
}

func ValidateListUnsubscribe(value string) (err error) {
	methods := strings.Split(value, ",")
	for _, method := range methods {
		method = strings.TrimSpace(method)
		if method == "" {
			return errors.New("empty unsubscribe method")
		}
		if method[0] != '<' {
			return errors.New("unsubscribe method must open with a left angle brace")
		}
		if method[len(method)-1] != '>' {
			return errors.New("unsubscribe method must close with a right angle brace")
		}
		if strings.HasPrefix(method, "<mailto:") {
			emailAddress, _, _ := strings.Cut(method[8:], "?")
			if err = address.ValidateFormat(emailAddress); err != nil {
				return fmt.Errorf("invalid address format for list unsubscribe contact %s: %w", method, err)
			}
		} else if _, err = url.Parse(method[1 : len(method)-1]); err != nil {
			return fmt.Errorf("invalid list unbscribe URL %s: %w", method, err)
		}
	}
	return nil
}

func ValidateListUnsubscribePost(value string) (err error) {
	key, postValue, ok := strings.Cut(value, "=")
	if !ok || strings.TrimSpace(postValue) == "" {
		return fmt.Errorf("unsubscribe post header does not include a value after equals sign: %s", value)
	}
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("unsubscribe post header does not include a key before equals sign: %s", value)
	}
	return nil
}
