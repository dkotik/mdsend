package media

import (
	"bytes"
	"image"
	"image/png"
)

func EncodePNG(image image.Image) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := png.Encode(b, image); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
