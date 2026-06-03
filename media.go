package mdsend

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/dkotik/mdsend/internal/mime"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

type MediaConstraints struct {
	Width   int
	Height  int
	Quality int
}

func (m MediaConstraints) WithResolution(resolution int) MediaConstraints {
	const resolutionRatio = float32(1920 / 1080)
	return MediaConstraints{
		Width:   resolution,
		Height:  int(float32(resolution) * resolutionRatio),
		Quality: m.Quality,
	}
}

func (l Letter) GetMediaConstraints() (m MediaConstraints, err error) {
	switch media := l.Frontmatter[FieldNameMediaContraints].(type) {
	case nil:
		return MediaConstraints{}, nil
	case map[string]any:
		m.Quality, err = getPercentageFromMap(media, FieldNameMediaQuality, 80)
		if err != nil {
			return m, err
		}
		resolution, err := getIntFromMap(media, FieldNameMediaResolution, 1080)
		if err != nil {
			return m, err
		}
		if resolution < 160 {
			return m, errors.New("resolution must be at least 160")
		}
		if resolution > 7680 {
			return m, fmt.Errorf("resolution must be at most 7680")
		}
		m = m.WithResolution(resolution)
		m.Width, err = getIntFromMap(media, FieldNameMediaWidth, m.Width)
		if err != nil {
			return m, err
		}
		m.Height, err = getIntFromMap(media, FieldNameMediaHeight, m.Height)
		if err != nil {
			return m, err
		}
		return m, nil
	default:
		return MediaConstraints{}, fmt.Errorf("invalid media constraints %T: %v", media, media)
	}
}

func (m MediaConstraints) Compress(image image.Image) image.Image {
	return resize.Thumbnail(
		uint(m.Width),
		uint(m.Height),
		image,
		resize.Lanczos3,
	)
}

func (m MediaConstraints) EncodeJPEG(image image.Image) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := jpeg.Encode(b, image, &jpeg.Options{
		Quality: m.Quality,
	}); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (m MediaConstraints) EncodePNG(image image.Image) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := png.Encode(b, image); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (m MediaConstraints) Apply(a Attachment) (_ Attachment, err error) {
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
		a.Content, err = m.EncodeJPEG(m.Compress(image))
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
		a.Content, err = m.EncodePNG(m.Compress(image))
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
		a.Content, err = m.EncodePNG(m.Compress(image))
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
		a.Content, err = m.EncodePNG(m.Compress(image))
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
		a.Content, err = m.EncodeJPEG(m.Compress(image))
		if err != nil {
			return a, err
		}
		a.ContentType = mime.ContentTypeImageJPEG
		return a, nil
	default:
		return a, nil
	}
}

func estimateJPEGQuality(r io.Reader) (int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	// Ensure it's a JPEG
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 0, fmt.Errorf("not a valid JPEG file")
	}

	idx := 2
	for idx < len(data)-1 {
		if data[idx] != 0xFF {
			idx++
			continue
		}
		marker := data[idx+1]
		if marker == 0xD9 { // End of Image
			break
		}
		// DQT Marker is 0xDB
		if marker == 0xDB {
			length := int(data[idx+2])<<8 + int(data[idx+3])
			if idx+2+length > len(data) {
				break
			}
			tableData := data[idx+4 : idx+2+length]
			return calculateQualityFromDQT(tableData), nil
		}
		idx += 2
	}
	return 0, fmt.Errorf("could not estimate quality: DQT marker missing")
}

func calculateQualityFromDQT(table []byte) int {
	// Fallback to checking the first few coefficients against IJG standards
	// A simple heuristic based on the first luminance AC coefficient:
	if len(table) < 3 {
		return 95
	}
	firstCoeff := int(table[1]) // Value at index 1 of the luma table
	if firstCoeff == 0 {
		return 100
	}
	// Approximate reverse-engineered IJG formula
	quality := 100 - firstCoeff
	if quality < 1 {
		return 1
	}
	return quality
}

// deprecate everything below =====================================================

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
