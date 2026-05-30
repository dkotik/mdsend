/*
Package mime encodes electronic mail parts for delivery.
*/
package mime

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/textproto"
	"sort"
)

const (
	// LineLengthLimit is the maximum length of a line
	// for MIME encoding.
	//
	// RFC 5322 2.1.1 limits to 78, excluding CRLF. mime/quotedprintable sets this to 76.
	LineLengthLimit = 76

	ContentTypeImageJPEG = "image/jpeg"
	ContentTypeImagePNG  = "image/png"
	ContentTypeImageGIF  = "image/gif"
	// ContentTypeImageWEBP = ""
)

func WriteHeader(w io.Writer, header textproto.MIMEHeader) (err error) {
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

// NewEncoderBase64 encodes data to standard Base64 encoding
// while keeping each line under [LineLengthLimit].
// Encoder must be closed to work correctly.
func NewEncoderBase64(w io.Writer) io.WriteCloser {
	return base64.NewEncoder(base64.StdEncoding, &lineWrapper{w: w})
}

// lineWrapper writes everything to io.Writer chunked by [LineLengthLimit].
type lineWrapper struct {
	w     io.Writer
	wrote int
}

func (w *lineWrapper) Write(p []byte) (int, error) {
	leftover := w.wrote % LineLengthLimit
	lineSize := LineLengthLimit - leftover
	for i := 0; i < len(p); i += lineSize {
		// Reset linesize
		if i%LineLengthLimit != 0 && lineSize < LineLengthLimit {
			lineSize = LineLengthLimit
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
