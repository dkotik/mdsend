package renderers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/http"
	"net/textproto"
	"path/filepath"
	"sort"

	"github.com/dkotik/mdsend/loaders"
)

// Renderer generates HTML and alternative email bodies to be used by Provider.
type Renderer interface {
	Render(w io.Writer, message *loaders.Message, to string) error
}

func writeHeader(w io.Writer, header textproto.MIMEHeader) {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range header[k] {
			fmt.Fprintf(w, "%s: %s\r\n", k, v)
		}
	}
	fmt.Fprintf(w, "\r\n")
}

func writeTextPart(mime *multipart.Writer, source io.Reader, isHTML bool) error {
	header := textproto.MIMEHeader{
		`Content-Transfer-Encoding`: {`quoted-printable`}, `Mime-Version`: {`1.0`}}
	if isHTML {
		header.Set(`Content-Type`, `text/html; charset=utf-8`)
	} else {
		header.Set(`Content-Type`, `text/plain; charset=utf-8`)
	}
	w, err := mime.CreatePart(header)
	if err != nil {
		return err
	}
	q := quotedprintable.NewWriter(w)
	defer q.Close()
	_, err = io.Copy(q, source)
	return err
}

func writeTextPartsAndInlineFiles(
	w io.Writer, mime *multipart.Writer, plain []byte, HTML []byte) error {
	if err := writeTextPart(mime, bytes.NewReader(plain), false); err != nil {
		return err
	}
	inline, HTML := InlineImages(HTML)
	if len(inline) == 0 { // no need to nest multipart/related
		return writeTextPart(mime, bytes.NewReader(HTML), true)
	}
	nested := multipart.NewWriter(w)
	defer nested.Close()
	_, err := mime.CreatePart(textproto.MIMEHeader{`Content-Type`: {fmt.Sprintf(`multipart/related; boundary="%s"`, nested.Boundary())}})
	if err != nil {
		return err
	}
	if err = writeTextPart(nested, bytes.NewReader(HTML), true); err != nil {
		return err
	}
	for _, file := range inline {
		if err = mimeAttachment(file, true, nested); err != nil {
			return err
		}
	}
	return err
}

func mimeAttachment(file string, inline bool, mime *multipart.Writer) error {
	// handle, err := os.Open(file)
	// Don't go over 600px in width in email layout.
	handle, err := resizedReadCloser(file, 600, 400, 75)
	// handle, err := resizedReadCloser(file, 4, 4, 4)
	if err != nil {
		return err
	}
	defer handle.Close()

	// Determine content type, read bytes will be written out below.
	fileHead := make([]byte, 512)
	fileHeadBytesRead, err := handle.Read(fileHead)
	if err != nil && err != io.EOF {
		return err
	}

	header := textproto.MIMEHeader{}
	header.Set(`Content-Type`, http.DetectContentType(fileHead[:fileHeadBytesRead]))
	fname := filepath.Base(file)
	if inline {
		header.Set(`Content-Disposition`, fmt.Sprintf(`inline; filename=%q`, fname))
		header.Set(`Content-ID`, fmt.Sprintf(`<%s>`, fname))
	} else {
		header.Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename=%q`, fname))
	}
	header.Set(`Content-Transfer-Encoding`, `base64`)
	partWriter, err := mime.CreatePart(header)
	if err != nil {
		return err
	}
	w := base64.NewEncoder(base64.StdEncoding, &MimeLineWrapper{w: partWriter})
	defer w.Close()
	if _, err2 := w.Write(fileHead[:fileHeadBytesRead]); err2 != nil {
		return err2
	}
	_, err = io.Copy(w, handle)
	return err
}
