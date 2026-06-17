package renderers

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png" // Allow resizing PNG images.
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nfnt/resize"
)

// RFC 5322 2.1.1 limits to 78, excluding CRLF. mime/quotedprintable sets this to 76.
const maxMimeLineLen = 76
const resizableFormats = `.jpg.jpeg.png`

var reInlineImages = regexp.MustCompile(`\<img[^\>]+src\=\"([^\"]+)`)

// InlineImages returns a list of detected inline image paths
// and replaces the src paths with "cid:filename" tag.
func InlineImages(b []byte) ([]string, []byte) {
	result := make([]string, 0)
	return result, reInlineImages.ReplaceAllFunc(b, func(m []byte) []byte {
		// log.Fatal(spew.Sdump(m), index)
		index := bytes.LastIndexByte(m, '"') + 1
		p := string(m[index:])
		result = append(result, p)
		return append(m[:index], append([]byte(`cid:`), []byte(filepath.Base(p))...)...)
		// return append(m[:index], []byte(`cid:yepx`)...)
	})
}

func resizedReadCloser(file string, maxWidth, maxHeight uint, quality int) (io.ReadCloser, error) {
	handle, err := os.Open(file)
	if err != nil {
		return handle, err
	}
	// defer handle.Close()
	if strings.Index(resizableFormats, strings.ToLower(filepath.Ext(file))) == -1 {
		return handle, err
	}
	m, _, err := image.Decode(handle)
	if err != nil {
		// panic(file)
		return handle, err
	}
	t := resize.Thumbnail(maxWidth, maxHeight, m, resize.Lanczos3)
	buffer := bytes.NewBuffer(nil)
	err = jpeg.Encode(buffer, t, &jpeg.Options{Quality: quality})
	if err != nil {
		return handle, err
	}
	return ioutil.NopCloser(buffer), nil
}

// MimeLineWrapper writes everything to io.Writer chunked by maxMimeLineLen.
type MimeLineWrapper struct {
	w     io.Writer
	wrote int
}

func (w *MimeLineWrapper) Write(p []byte) (int, error) {
	leftover := w.wrote % maxMimeLineLen
	lineSize := maxMimeLineLen - leftover
	for i := 0; i < len(p); i += lineSize {
		// Reset linesize
		if i%maxMimeLineLen != 0 && lineSize < maxMimeLineLen {
			lineSize = maxMimeLineLen
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
