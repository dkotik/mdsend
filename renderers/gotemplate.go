package renderers

// Helpful to debugging:
// - https://github.com/ohwgiles/wemed
// - this was helpful https://play.golang.org/p/Ifztb4dKFW2

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/dkotik/mdsend/loaders"

	blackfriday "github.com/russross/blackfriday/v2"
)

var goDefaultTemplate = template.Must(template.New(`default`).Parse(`
	<!doctype html>
	<html>
	  <head>
	    <meta name="viewport" content="width=device-width" />
		<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
		<title>{{ .Subject }}</title>
	  </head>
	  <body style="color: #1d1d1d;">
	  	<section style="background-color: #e8ecfb; padding: 0.5rem 1rem 0.5rem 1.5rem;">
		  {{ .Data.content }}
		</section>
	  </body>
	</html>`))

// GoTemplateMIMERenderer renders map to MIME buffer.
type GoTemplateMIMERenderer struct {
	body []byte
}

func (r *GoTemplateMIMERenderer) renderMarkdown(w io.Writer, m *loaders.Message) error {
	m.Data[`content`] = template.HTML(blackfriday.Run(r.body))
	stack := make([]string, 0)
	switch tmpl := m.Data[`template`].(type) {
	case []interface{}: // a list of templates
		for _, one := range tmpl {
			stack = append(stack, fmt.Sprintf(`%v`, one))
		}
	case string: // one template
		stack = append(stack, tmpl)
	}
	if len(stack) > 0 {
		t, err := template.ParseFiles(stack...)
		if err != nil {
			return fmt.Errorf(`could not fetch template: %w`, err)
		}
		return t.Execute(w, m)
	}
	return goDefaultTemplate.Execute(w, m)
}

// Render returns MIME encoded buffer.
func (r *GoTemplateMIMERenderer) Render(w io.Writer, m *loaders.Message, to string) error {
	mime := multipart.NewWriter(w)
	defer mime.Close()
	header := m.MIMEHeader() // this can be optimized, since header does not change between emails being sent
	header.Set("To", to)

	// https://mailtrap.io/blog/list-unsubscribe-header/
	var unsubscribeLinks []string
	if m.UnsubscribeLink != nil {
		// m.UnsubscribeLink.Query().Set("address", to) // on copy operation
		link := fmt.Sprintf(`%s?address=%s&list=%s`,
			m.UnsubscribeLink.String(),
			url.QueryEscape(to),
			url.QueryEscape(m.ListID))
		m.Data["injectedUnsubscribeLink"] = link
		unsubscribeLinks = append(unsubscribeLinks, fmt.Sprintf(`<%s>`, link))
	}
	if m.UnsubscribeContact != nil {
		unsubscribeLinks = append(unsubscribeLinks,
			fmt.Sprintf(`<mailto: %s?subject=unsubscribe>`, m.UnsubscribeContact.Email))
	}
	if len(unsubscribeLinks) > 0 {
		if m.ListID == "" {
			return errors.New("ListID is required for unsubscribing")
		}
		header.Set("List-ID", m.ListID)
		header.Set("List-Unsubscribe", strings.Join(unsubscribeLinks, `, `))
		header.Set("List-Unsubscribe-Post", "List-Unsubscribe=One-Click")
	}

	// render body with template values
	bf := bytes.NewBuffer(nil)
	err := template.Must(template.New(`body`).Parse(string(m.Body))).Execute(bf, m)
	if err != nil {
		return err
	}
	r.body = bf.Bytes()
	// render markdown
	b := bytes.NewBuffer(nil)
	if err = r.renderMarkdown(b, m); err != nil {
		return err
	}

	if len(m.Attachments) == 0 {
		header.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=\"%s\"; charset=\"utf-8\"", mime.Boundary()))
		writeHeader(w, header)
		return writeTextPartsAndInlineFiles(w, mime, bf.Bytes(), b.Bytes())
	}

	// have to nest into a mixed MIME
	header.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=\"%s\"; charset=\"utf-8\"", mime.Boundary()))
	writeHeader(w, header)
	nested := multipart.NewWriter(w)
	_, err = mime.CreatePart(textproto.MIMEHeader{`Content-Type`: {fmt.Sprintf(`multipart/alternative; boundary="%s"`, nested.Boundary())}})
	if err != nil {
		return err
	}
	if err = writeTextPartsAndInlineFiles(w, nested, bf.Bytes(), b.Bytes()); err != nil {
		return err
	}
	nested.Close()

	for _, file := range m.Attachments {
		if err = mimeAttachment(file, false, mime); err != nil {
			return err
		}
	}
	return err
}
