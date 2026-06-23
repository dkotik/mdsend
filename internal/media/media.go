package media

import (
	"bytes"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"

	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

const (
	// TODO: those are duplicates of mime types in mime package because of cyclical imports: fix
	ContentTypeOctetStream = "application/octet-stream"
	ContentTypeTextPlain   = "text/plain"
	ContentTypeTextHTML    = "text/html"
	ContentTypeImageJPEG   = "image/jpeg"
	ContentTypeImageBMP    = "image/bmp"
	ContentTypeImagePNG    = "image/png"
	ContentTypeImageGIF    = "image/gif"
	ContentTypeImageWEBP   = "image/webp"
	ContentTypeZip         = "application/zip"
)

type Constraints struct {
	Width   int
	Height  int
	Quality int
}

func (m Constraints) WithResolution(resolution int) Constraints {
	const resolutionRatio = float32(1920 / 1080)
	return Constraints{
		Width:   resolution,
		Height:  int(float32(resolution) * resolutionRatio),
		Quality: m.Quality,
	}
}

func (m Constraints) ApplyTo(b []byte) (_ []byte, contentType string, err error) {
	contentType = http.DetectContentType(b) // always returns a valid MIME type
	switch contentType {
	case ContentTypeOctetStream:
		return b, contentType, nil // default
	case ContentTypeImageJPEG:
		config, err := jpeg.DecodeConfig(bytes.NewReader(b))
		if err != nil {
			return b, contentType, err
		}
		if config.Width <= m.Width && config.Height <= m.Height {
			q, err := estimateJPEGQuality(bytes.NewReader(b))
			if err == nil && q <= m.Quality {
				return b, contentType, nil // no need to resize
			}
		}
		image, err := jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			return b, contentType, err
		}
		b, err = EncodeJPEG(Resize(image, uint(m.Width), uint(m.Height)), m.Quality)
		return b, contentType, err
	case ContentTypeImagePNG:
		config, err := png.DecodeConfig(bytes.NewReader(b))
		if err != nil {
			return b, contentType, err
		}
		if config.Width <= m.Width && config.Height <= m.Height {
			return b, contentType, nil // no need to resize
		}
		image, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			return b, contentType, err
		}
		b, err = EncodePNG(Resize(image, uint(m.Width), uint(m.Height)))
		if err != nil {
			return b, contentType, err
		}
		return b, ContentTypeImagePNG, nil
	case ContentTypeImageWEBP:
		image, err := webp.Decode(bytes.NewReader(b))
		if err != nil {
			return b, contentType, err
		}
		b, err = EncodePNG(Resize(image, uint(m.Width), uint(m.Height)))
		if err != nil {
			return b, contentType, err
		}
		return b, ContentTypeImagePNG, nil
	case ContentTypeImageGIF:
		image, err := gif.Decode(bytes.NewReader(b))
		if err != nil {
			return b, contentType, err
		}
		b, err = EncodePNG(Resize(image, uint(m.Width), uint(m.Height)))
		if err != nil {
			return b, contentType, err
		}
		return b, ContentTypeImagePNG, nil
	case ContentTypeImageBMP:
		image, err := bmp.Decode(bytes.NewReader(b))
		if err != nil {
			return b, contentType, err
		}
		b, err = EncodeJPEG(Resize(image, uint(m.Width), uint(m.Height)), m.Quality)
		if err != nil {
			return b, contentType, err
		}
		return b, ContentTypeImageJPEG, nil
	default:
		return b, contentType, nil
	}
}
