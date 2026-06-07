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
)

const (
	HeaderMIMEVersion             = "MIME-Version"
	HeaderContentType             = "Content-Type"
	HeaderContentTransferEncoding = "Content-Transfer-Encoding"
	HeaderContentID               = "Content-ID"
	HeaderContentDescription      = "Content-Description"
	HeaderContentDisposition      = "Content-Disposition"
	HeaderDate                    = "Date"
)

// taken from mime package of the standard library
func needsEncoding(s string) bool {
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
	i, err = w.Write([]byte("\r\n"))
	n += i
	if err != nil {
		return n, err
	}
	if i != 2 {
		return n, io.ErrShortWrite
	}
	return n, nil
}

var maxBase64Len = base64.StdEncoding.DecodedLen(WordLengthLimit)

// TODO: remove <n int> from WriteHeader signature
func WriteHeader(w io.Writer, name, value string) (n int, err error) {
	if !needsEncoding(value) {
		return writeSimpleHeader(w, name, value)
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
	var currentLen, last, runeLen int
	currentLen = len(name) - 2 // subtract the length of the header name and separator

	for i = 0; i < len(value); i += runeLen {
		// Multi-byte characters must not be split across encoded-words.
		// See RFC 2047, section 5.3.
		_, runeLen = utf8.DecodeRuneInString(value[i:])

		if currentLen+runeLen <= maxBase64Len {
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
			last = i
			currentLen = runeLen
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

// func (h Header) WriteTo(w io.Writer) (n int64, err error) {
// 	b := bufio.NewWriter(w)
// 	i, err := b.WriteString(h.Name)
// 	n = int64(i)
// 	if err != nil {
// 		return n, err
// 	}
// 	if err = b.WriteByte(':'); err != nil {
// 		return n, err
// 	}
// 	if err = b.WriteByte(' '); err != nil {
// 		return n + 1, err
// 	}
// 	n += 4 // TODO: why 4 instead of 2?
// 	if needsEncoding(h.Value) {
// 		// panic("UTF8 encoding not supported")
// 		i, err := b.WriteString(mime.BEncoding.Encode("utf-8", h.Value))
// 		n += int64(i)
// 		if err != nil {
// 			return n, err
// 		}
// 		return n, b.Flush()
// 	}
// 	lastIndex := len(h.Value) - 1
// 	for i, c := range h.Value {
// 		if n%LineLengthLimit == 0 && i < lastIndex {
// 			// Write the CRLF and the continuation space
// 			if err = b.WriteByte('\r'); err != nil {
// 				return n, err
// 			}
// 			if err = b.WriteByte('\n'); err != nil {
// 				return n + 1, err
// 			}
// 			if err = b.WriteByte(' '); err != nil {
// 				return n + 2, err
// 			}
// 			n += 3
// 		}
// 		if err = b.WriteByte(byte(c)); err != nil {
// 			return n, err
// 		}
// 		n++
// 	}
// 	if err = b.WriteByte('\r'); err != nil {
// 		return n, err
// 	}
// 	if err = b.WriteByte('\n'); err != nil {
// 		return n + 1, err
// 	}
// 	return n + 2, b.Flush()
// }

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

// func WriteHeaders(w io.Writer, hs ...Header) (err error) {
// 	// if _, err = fmt.Fprint(w, HeaderMIMEVersion+": 1.0\r\n"); err != nil {
// 	// 	return err
// 	// }

// 	_, err = fmt.Fprintf(w, "\r\n")
// 	return err
// }

// // DEPRECATED: use WriteHeaders instead
// func WriteHeader(w io.Writer, header textproto.MIMEHeader) (err error) {
// 	// sourced from go/src/mime/multipart/writer.go
// 	keys := make([]string, 0, len(header))
// 	for k := range header {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)
// 	for _, k := range keys {
// 		for _, v := range header[k] {
// 			if _, err = fmt.Fprintf(w, "%s: %s\r\n", k, v); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	if _, err = fmt.Fprintf(w, "\r\n"); err != nil {
// 		return err
// 	}
// 	return nil
// }
