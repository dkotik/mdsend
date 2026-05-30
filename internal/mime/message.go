package mime

import (
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
)

type Message struct {
	Entropy             rand.Source
	Header              textproto.MIMEHeader
	Text                string
	HTML                string
	Attachments         []Attachment
	EmbeddedAttachments []EmbeddedAttachment
}

func (m Message) WriteText(mw *MultipartWriter) (err error) {
	header := textproto.MIMEHeader{
		`Content-Transfer-Encoding`: {`quoted-printable`}, `Mime-Version`: {`1.0`}}
	header.Set(`Content-Type`, `text/plain; charset=utf-8`)
	w, err := mw.CreatePart(header)
	if err != nil {
		return err
	}
	q := quotedprintable.NewWriter(w)
	_, err = io.Copy(q, strings.NewReader(m.Text))
	return errors.Join(err, q.Close())
}

func (m Message) WriteHTML(mw *MultipartWriter) (err error) {
	header := textproto.MIMEHeader{
		`Content-Transfer-Encoding`: {`quoted-printable`}, `Mime-Version`: {`1.0`}}
	header.Set(`Content-Type`, `text/html; charset=utf-8`)
	w, err := mw.CreatePart(header)
	if err != nil {
		return err
	}
	q := quotedprintable.NewWriter(w)
	_, err = io.Copy(q, strings.NewReader(m.HTML))
	return errors.Join(err, q.Close())
}

func (m Message) Encode(w io.Writer) (err error) {
	mime := NewMultipartWriter(w, fmt.Sprintf("%x", m.Entropy.Uint64()))
	defer func() {
		err = errors.Join(err, mime.Close())
	}()

	if len(m.Attachments)+len(m.EmbeddedAttachments) == 0 {
		m.Header.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=\"%s\"; charset=\"utf-8\"", mime.Boundary()))
		if err = writeHeader(w, m.Header); err != nil {
			return err
		}
		if err = m.WriteText(mime); err != nil {
			return err
		}
		if err = m.WriteHTML(mime); err != nil {
			return err
		}
	}

	// have to nest into a mixed MIME
	m.Header.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=\"%s\"; charset=\"utf-8\"", mime.Boundary()))
	writeHeader(w, m.Header)
	nested := NewMultipartWriter(w, fmt.Sprintf("%x", m.Entropy.Uint64()))
	defer func() {
		err = errors.Join(err, nested.Close())
	}()
	_, err = mime.CreatePart(textproto.MIMEHeader{`Content-Type`: {fmt.Sprintf(`multipart/alternative; boundary="%s"`, nested.Boundary())}})
	if err != nil {
		return err
	}
	if err = m.WriteText(mime); err != nil {
		return err
	}
	if err = m.WriteHTML(mime); err != nil {
		return err
	}
	for _, file := range m.Attachments {
		if err = file.Encode(nested); err != nil {
			return err
		}
	}

	return err
}
