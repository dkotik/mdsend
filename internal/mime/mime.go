/*
Package mime encodes electronic mail parts for delivery.
*/
package mime

import (
	"encoding/base64"
	"io"
	"math/rand/v2"
)

const (
	// LineLengthLimit is the maximum length of a line
	// for MIME encoding.
	//
	// RFC 5322 2.1.1 limits to 78, excluding CRLF. mime/quotedprintable sets this to 76.
	LineLengthLimit = 76

	// BoundaryLengthLimit is the maximum length of a boundary string
	// for MIME encoding according to RFC 1341 set to 70 characters. Boundary can never
	// wrap to the next line, as space characters are not allowed in boundaries.
	// Therefore, the maximum length must be reduced by the length of the parameter name.
	BoundaryLengthLimit = 70 - len(`boundary`) + 2

	// Mime package BEncoder refers to RFC 2047, section 2 to set
	// maximum word length to 75 characters. from which the length
	// of the prefix and suffix are subtracted to get the limit.
	// One less is to prevent line wrapping in the middle of a multi-byte
	// UTF character rune.
	WordLengthLimit = LineLengthLimit - 1 - len("=?UTF-8?b?") - len("?=")

	BEncodingPrefix = "=?UTF-8?b?"
	BEncodingSuffix = "?="
	CRNL            = "\r\n"

	ContentTypeTextPlain = "text/plain"
	ContentTypeTextHTML  = "text/html"
	ContentTypeImageJPEG = "image/jpeg"
	ContentTypeImageBMP  = "image/bmp"
	ContentTypeImagePNG  = "image/png"
	ContentTypeImageGIF  = "image/gif"
	ContentTypeImageWEBP = "image/webp"
	ContentTypeZip       = "application/zip"
)

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

const (
	boundaryCharset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'()+_,-./:=?"
	boundaryCharsetLength = len(boundaryCharset)
)

func NewBoundary(r *rand.Rand) string {
	b := make([]byte, BoundaryLengthLimit)
	for i := range b {
		b[i] = boundaryCharset[r.IntN(boundaryCharsetLength)]
	}
	return string(b)
}
