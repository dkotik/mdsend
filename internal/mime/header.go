package mime

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/textproto"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/dkotik/mdsend/header"
)

// taken from mime package of the standard library
func doesValueRequireEncoding(s string) bool {
	for _, b := range s {
		if (b < ' ' || b > '~') && b != '\t' {
			return true
		}
	}
	return false
}

func writeSimpleHeader(w io.Writer, name, value string) (n int, err error) {
	i, err := w.Write([]byte(name))
	if err != nil {
		return i, err
	}
	if i != len(name) {
		return n, io.ErrShortWrite
	}
	n = i
	i, err = w.Write([]byte(": "))
	n += i
	if err != nil {
		return n, err
	}
	if i != 2 {
		return n, io.ErrShortWrite
	}

	b := []byte(value)
	remaining := len(b)
	limit := min(remaining, LineLengthLimit-n-2)
	i, err = w.Write(b[:limit])
	n += i
	if err != nil {
		return n, err
	}
	if i == remaining {
		goto close // all bytes written in the first line
	}
	if i != limit {
		return n, io.ErrShortWrite
	}
	b = b[i:]
	remaining -= i

	for remaining > 0 {
		// Write the CRLF and the continuation space
		i, err = w.Write([]byte(CRNL + " "))
		n += i
		if err != nil {
			return n, err
		}
		if i != 3 {
			return n, io.ErrShortWrite
		}
		limit = min(remaining, LineLengthLimit-3)
		i, err = w.Write(b[:limit])
		n += i
		if err != nil {
			return n, err
		}
		if i != limit {
			return n, io.ErrShortWrite
		}
		b = b[i:]
		remaining -= i
	}

close:
	i, err = w.Write([]byte(CRNL))
	n += i
	if err != nil {
		return n, err
	}
	if i != 2 {
		return n, io.ErrShortWrite
	}
	return n, nil
}

// Mime package BEncoder refers to RFC 2047, section 2 to set
// maximum word length to 75 characters. from which the length
// of the prefix and suffix are subtracted to get the limit.
// Plus one for the leading space.
var maxHeaderBase64Len = base64.StdEncoding.DecodedLen(
	75 - len(BEncodingPrefix) - len(BEncodingSuffix),
)

// TODO: remove <n int> from WriteHeader signature
func WriteHeader(w io.Writer, name, value string) (n int, err error) {
	if !doesValueRequireEncoding(value) {
		// Important: long lines will be wrapped, which can
		// break URLs because plain multi-line header decoder
		// will take `\r\n ` as a word break.
		//
		// [doesValueRequireEncoding] returns true for line breaks,
		// but it will not see those in long values, because the header
		// is not yet wrapped. Therefore, encoding must be coerced
		// on long values.
		//
		// The standard library `mime/encodedword.go` file includes the following
		// comment: "White-space and newline  characters separating two
		// encoded-words must be deleted." This means that line splits
		// will heal when decoding the header, preserving URLs and long values.
		//
		// There is a test condition that uses standard library word decoder
		// to ensure that decoded header matches the expected value.
		if len(name)+2+len(value) < LineLengthLimit {
			return writeSimpleHeader(w, name, value)
		}
	}

	i, err := w.Write([]byte(name))
	if err != nil {
		return i, err
	}
	if i != len(name) {
		return n, io.ErrShortWrite
	}
	n = i
	i, err = w.Write([]byte(": " + BEncodingPrefix))
	n += i
	if err != nil {
		return n, err
	}
	if i != 12 {
		return n, io.ErrShortWrite
	}

	w64 := base64.NewEncoder(base64.StdEncoding, w)
	var (
		currentLen = n - i // len(": ") + space
		// currentLen = n
		runeLen int
		last    int
	)

	for i = 0; i < len(value); i += runeLen {
		// Multi-byte characters must not be split across encoded-words.
		// See RFC 2047, section 5.3.
		_, runeLen = utf8.DecodeRuneInString(value[i:])

		if currentLen+runeLen <= maxHeaderBase64Len {
			currentLen += runeLen
		} else {
			io.WriteString(w64, value[last:i])
			w64.Close()
			_, err = w.Write([]byte( // split word
				BEncodingSuffix + CRNL + " " + BEncodingPrefix,
			))
			// n += _
			if err != nil {
				return n, err
			}
			// if i != 12 {
			// 	return n, io.ErrShortWrite
			// }
			currentLen = runeLen
			last = i
		}
	}
	io.WriteString(w64, value[last:])
	w64.Close()

	i, err = w.Write([]byte(BEncodingSuffix + CRNL))
	n += i
	if err != nil {
		return n, err
	}
	if i != 4 {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func WriteTimeHeader(w io.Writer, name string, t time.Time) (int, error) {
	// mime.Header{Name: HeaderDate, Value: t.Format(time.RFC1123Z)}
	return WriteHeader(w, name, t.Format(time.RFC1123Z))
}

func newHeaderTemplate(header textproto.MIMEHeader) (t *template.Template, err error) {
	b := &bytes.Buffer{}
	if err = writeHeader(b, header); err != nil {
		return nil, err
	}
	t, err = template.New("").Parse(b.String())
	if err != nil {
		return nil, fmt.Errorf("invalid header template: %w", err)
	}
	return t, nil
}

func writeHeader(w io.Writer, header textproto.MIMEHeader) (err error) {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range header[k] {
			if _, err = fmt.Fprintf(w, "%s: %s\r\n", k, v); err != nil {
				return err
			}
		}
	}
	_, err = fmt.Fprintf(w, "\r\n")
	return err
}

// func (w Writer) WriteAlternativeBoundaryHeader(out io.Writer) (err error) {
// 	if _, err = out.Write([]byte(HeaderContentType + `: multipart/alternative;` + CRNL + `boundary="`)); err != nil {
// 		return err
// 	}
// 	if _, err = out.Write([]byte(w.mixedBoundary)); err != nil {
// 		return err
// 	}
// 	if _, err = out.Write([]byte(`"` + CRNL)); err != nil {
// 		return err
// 	}
// 	return nil
// }

func (w Writer) WriteMixedBoundaryHeader(out io.Writer) (err error) {
	if _, err = out.Write([]byte(header.ContentType + `: multipart/mixed;` + CRNL + ` boundary="`)); err != nil {
		return err
	}
	if _, err = out.Write([]byte(w.mixedBoundary)); err != nil {
		return err
	}
	if _, err = out.Write([]byte(`"` + CRNL)); err != nil {
		return err
	}
	return nil
}

func (w Writer) WriteRelatedBoundaryHeader(out io.Writer) (err error) {
	if _, err = out.Write([]byte(header.ContentType + `: multipart/related;` + CRNL + ` boundary="`)); err != nil {
		return err
	}
	if _, err = out.Write([]byte(w.relatedBoundary)); err != nil {
		return err
	}
	if _, err = out.Write([]byte(`"` + CRNL)); err != nil {
		return err
	}
	return nil
}
