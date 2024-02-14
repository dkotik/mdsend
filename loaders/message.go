package loaders

import (
	"fmt"
	"net/textproto"
	"net/url"
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
	BCC                *[]Participant
	UnsubscribeContact *Participant
	UnsubscribeLink    *url.URL
	Attachments        []string
	Body               []byte
	Data               map[string]interface{}
}

// MIMEHeader provides MIME fields for the renderer.
func (m *Message) MIMEHeader() textproto.MIMEHeader {
	h := textproto.MIMEHeader{
		"Mime-Version": {"1.0"},
		"Subject":      {m.Subject},
		"From":         {m.From},
	}
	if len(m.ReplyTo) > 0 {
		h.Set("Reply-To", m.ReplyTo)
		// } else {
		// 	h.Set("Reply-To", m.From)
	}
	if len(m.Comments) > 0 {
		h.Set("Comments", m.Comments)
	}
	if len(m.Keywords) > 0 {
		h.Set("Keywords", m.Keywords)
	}
	return h
}
