package loaders

import (
	"errors"
	"fmt"
	"net/mail"
	"net/textproto"
	"regexp"
	"text/template"
)

// Participant is a person or agent in "to", "from", "cc", and "bcc" fields.
type Participant struct {
	Name   string
	Email  string
	Source string
	Data   map[string]interface{}
}

func (p Participant) String() string {
	if p.Name == `` {
		return p.Email
	}
	return fmt.Sprintf(`%s <%s>`, p.Name, p.Email)
	// email := p.Data[`email`].(string)
	// if name := p.Data[`name`].(string); len(name) > 0 {
	// 	return fmt.Sprintf(`%s <%s>`, name, email)
	// }
	// return email
}

// Message models an email message.
type Message struct {
	Source             string // file from which the message is generated
	Current            *Participant
	Date               string
	Subject            string
	From               string
	ReplyTo            string
	Comments           string
	Keywords           string
	To                 *[]Participant
	CC                 *[]Participant
	ListID             string
	BCC                *[]Participant
	UnsubscribeContact *Participant
	UnsubscribeLink    *template.Template
	Attachments        []string
	Body               []byte
	Data               map[string]interface{}
}

// MIMEHeader provides MIME fields for the renderer.
func (m *Message) MIMEHeader() textproto.MIMEHeader {
	h := textproto.MIMEHeader{
		"MIME-Version": {"1.0"},
		"Subject":      {m.Subject},
		"From":         {m.From},
	}
	if len(m.ReplyTo) > 0 {
		h.Set("Reply-To", m.ReplyTo)
	}
	if len(m.Comments) > 0 {
		h.Set("Comments", m.Comments)
	}
	if len(m.Keywords) > 0 {
		h.Set("Keywords", m.Keywords)
	}
	return h
}

func (m *Message) SetListID(ID string) error {
	if !regexp.MustCompile(`[^\<\>]+\s+\<[\w\.]+\>`).MatchString(ID) {
		return fmt.Errorf("list ID does not match format `Description <subdomain.domain.com>`: %s", ID)
	}
	m.ListID = ID
	return nil
}

func (m *Message) SetUnsubscribeLink(URL string) (err error) {
	if m.ListID == "" {
		return errors.New("ListID is required for the unsubscribe link")
	}
	m.UnsubscribeLink, err = NewUnsubscribeLinkTemplate(URL)
	if err != nil {
		return fmt.Errorf("cannot parse unsubscribe link %q: %w", URL, err)
	}
	return nil
}

func (m *Message) SetUnsubscribeContact(address string) error {
	if m.ListID == "" {
		return errors.New("ListID is required for the unsubscribe contact")
	}
	a, err := mail.ParseAddress(address)
	if err != nil {
		return fmt.Errorf("cannot parse unsubscribeContact %q: %w", address, err)
	}
	m.UnsubscribeContact = &Participant{
		Name:  a.Name,
		Email: a.Address,
	}
	return nil
}
