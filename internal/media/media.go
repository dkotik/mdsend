package media

import (
	"bytes"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/mime"
	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

func Compress(a mdsend.Attachment, m mdsend.MediaConstraints) (_ mdsend.Attachment, err error) {
	switch a.ContentType {
	case mime.ContentTypeImageJPEG:
		config, err := jpeg.DecodeConfig(bytes.NewReader(a.Content))
		if err != nil {
			return a, err
		}
		if config.Width <= m.Width && config.Height <= m.Height {
			q, err := estimateJPEGQuality(bytes.NewReader(a.Content))
			if err == nil && q <= m.Quality {
				return a, nil // no need to resize
			}
		}
		image, err := jpeg.Decode(bytes.NewReader(a.Content))
		if err != nil {
			return a, err
		}
		a.Content, err = EncodeJPEG(Resize(image, uint(m.Width), uint(m.Height)), m.Quality)
		return a, err
	case mime.ContentTypeImagePNG:
		config, err := png.DecodeConfig(bytes.NewReader(a.Content))
		if err != nil {
			return a, err
		}
		if config.Width <= m.Width && config.Height <= m.Height {
			return a, nil // no need to resize
		}
		image, err := png.Decode(bytes.NewReader(a.Content))
		if err != nil {
			return a, err
		}
		a.Content, err = EncodePNG(Resize(image, uint(m.Width), uint(m.Height)))
		if err != nil {
			return a, err
		}
		a.ContentType = mime.ContentTypeImagePNG
		return a, nil
	case mime.ContentTypeImageWEBP:
		image, err := webp.Decode(bytes.NewReader(a.Content))
		if err != nil {
			return a, err
		}
		a.Content, err = EncodePNG(Resize(image, uint(m.Width), uint(m.Height)))
		if err != nil {
			return a, err
		}
		a.ContentType = mime.ContentTypeImagePNG
		return a, nil
	case mime.ContentTypeImageGIF:
		image, err := gif.Decode(bytes.NewReader(a.Content))
		if err != nil {
			return a, err
		}
		a.Content, err = EncodePNG(Resize(image, uint(m.Width), uint(m.Height)))
		if err != nil {
			return a, err
		}
		a.ContentType = mime.ContentTypeImagePNG
		return a, nil
	case mime.ContentTypeImageBMP:
		image, err := bmp.Decode(bytes.NewReader(a.Content))
		if err != nil {
			return a, err
		}
		a.Content, err = EncodeJPEG(Resize(image, uint(m.Width), uint(m.Height)), m.Quality)
		if err != nil {
			return a, err
		}
		a.ContentType = mime.ContentTypeImageJPEG
		return a, nil
	default:
		return a, nil
	}
}
