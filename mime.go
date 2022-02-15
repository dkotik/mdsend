package mdsend

import (
	"fmt"
	"io"
	"net/textproto"
	"sort"
	"encoding/base64"
)

// RFC 5322 2.1.1 limits to 78, excluding CRLF. mime/quotedprintable sets this to 76.
const maxMIMELineLen = 76

func MIMEHeaderTo(w io.Writer, header textproto.MIMEHeader) (err error) {
    // sourced from go/src/mime/multipart/writer.go
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
    if _, err = fmt.Fprintf(w, "\r\n"); err != nil {
        return err
    }
    return nil
}

func MIMEBase64To(w io.Writer, r io.Reader) (err error) {
    encoder := base64.NewEncoder(base64.StdEncoding, &MIMELineWrapper{w: w})
    _, err = io.Copy(encoder, r)
    return err
}

// MIMELineWrapper writes everything to io.Writer chunked by maxMIMELineLen.
type MIMELineWrapper struct {
	w     io.Writer
	wrote int
}

func (w *MIMELineWrapper) Write(p []byte) (int, error) {
	leftover := w.wrote % maxMIMELineLen
	lineSize := maxMIMELineLen - leftover
	for i := 0; i < len(p); i += lineSize {
		// Reset linesize
		if i%maxMIMELineLen != 0 && lineSize < maxMIMELineLen {
			lineSize = maxMIMELineLen
		}
		// Calculate the end of the chunk offset
		end := i + lineSize
		if end > len(p) {
			end = len(p)
		}
		// Slice chunk out of p
		chunk := p[i:end]
		// Increment the amount wrote so far by the chunk size
		w.wrote += len(chunk)
		// Write the chunk
		if n, err := w.w.Write(chunk); err != nil {
			return i + n, err
		}
		// If this finishes a line, add linebreaks
		if end == i+lineSize {
			if _, err := w.w.Write([]byte("\r\n")); err != nil {
				// If this errors, return the bytes wrote so far from the
				// caller's perspective (it is unaware newlines are being added)
				return i + len(chunk), err
			}
		}
	}

	return len(p), nil
}
