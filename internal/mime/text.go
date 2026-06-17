package mime

import (
	"bytes"
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

func writeHTML(w io.Writer, text string) (err error) {
	_, err = fmt.Fprintf(w, HeaderContentType+": text/html; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n")
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
	if _, err = w.Write([]byte(HeaderContentType + `: multipart/alternative;` + CRNL + ` boundary="`)); err != nil {
		return err
	}
	if _, err = w.Write([]byte(boundary)); err != nil {
		return err
	}
	if _, err = w.Write([]byte(`"` + CRNL + CRNL)); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "--%s\r\n", boundary)
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
	if err = writeHTML(w, html); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\r\n--%s--\r\n", boundary)
	if err != nil {
		return err
	}
	return nil
}

func (w Writer) writeAlternativeWithAttachments(out io.Writer, text, html, boundary string, inline []inlineAttachment) (err error) {
	if len(inline) == 0 {
		return writeAlternative(out, text, html, boundary)
	}
	if _, err = out.Write([]byte(HeaderContentType + `: multipart/alternative;` + CRNL + ` boundary="`)); err != nil {
		return err
	}
	if _, err = out.Write([]byte(w.textBoundary)); err != nil {
		return err
	}
	if _, err = out.Write([]byte(`"` + CRNL + CRNL)); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "--%s\r\n", w.textBoundary)
	if err != nil {
		return err
	}
	if err = writeText(out, text); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.textBoundary)
	if err != nil {
		return err
	}
	if err = w.WriteRelatedBoundaryHeader(out); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.relatedBoundary)
	if err != nil {
		return err
	}
	if err = writeHTML(out, html); err != nil {
		return err
	}
	for _, attachment := range inline {
		_, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.relatedBoundary)
		if err != nil {
			return err
		}
		if err = attachment.WriteInlineHeader(out, attachment.CanonicalContentID); err != nil {
			return err
		}
		data, ok := w.cachedAttachmentContents[attachment.Hash]
		if !ok {
			return fmt.Errorf("attachment content not found: %s", attachment.Hash)
		}
		if _, err = io.Copy(out, bytes.NewReader(data)); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(out, "\r\n--%s--\r\n", w.relatedBoundary)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "\r\n--%s--\r\n", w.textBoundary)
	return err
}
