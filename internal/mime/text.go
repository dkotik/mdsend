package mime

import (
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"
)

func writeText(w io.Writer, text string) (err error) {
	_, err = fmt.Fprintf(w, HeaderContentType+": text/plain; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n")
	if err != nil {
		return err
	}
	q := quotedprintable.NewWriter(w)
	_, err = io.Copy(q, strings.NewReader(text))
	if err != nil {
		return err
	}
	return q.Close()
}

func writeAlternative(w io.Writer, text, html, boundary string) (err error) {
	if _, err = WriteHeader(w, HeaderContentType, fmt.Sprintf("multipart/alternative; boundary=\"%s\"; charset=\"utf-8\"", boundary)); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\r\n--%s\r\n", boundary)
	if err != nil {
		return err
	}
	if err = writeText(w, text); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\r\n--%s\r\n", boundary)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, HeaderContentType+": text/html; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n")
	if err != nil {
		return err
	}
	q := quotedprintable.NewWriter(w)
	_, err = io.Copy(q, strings.NewReader(html))
	if err != nil {
		return err
	}
	if err = q.Close(); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\r\n--%s--\r\n", boundary)
	if err != nil {
		return err
	}
	return nil
}
