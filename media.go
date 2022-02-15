package mdsend

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/nfnt/resize"
)

type Compressor func(io.Reader) (*bytes.Buffer, error)

var compressorsByContentType = make(map[string]Compressor)

func RegisterCompressor(contentType string, resizer Compressor) {
	if resizer == nil {
		delete(compressorsByContentType, contentType)
		return
	}
	compressorsByContentType[contentType] = resizer
}

func NewImageJPEGResizer(maxWidth, maxHeight, quality uint) Compressor {
	q := int(quality)
	return func(r io.Reader) (_ *bytes.Buffer, err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("failed to resize image: %w", err)
			}
		}()

		m, _, err := image.Decode(r)
		if err != nil {
			return
		}
		t := resize.Thumbnail(maxWidth, maxHeight, m, resize.Lanczos3)

		var b bytes.Buffer
		err = jpeg.Encode(&b, t, &jpeg.Options{Quality: q})
		if err != nil {
			return
		}
		return &b, nil
	}
}
