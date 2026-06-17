package mdsend

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/nfnt/resize"
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
		m.Quality, err = getPercentageFromMap(media, FieldNameMediaConstraintsQuality, 80)
		if err != nil {
			return m, err
		}
		resolution, err := getIntFromMap(media, FieldNameMediaConstrainsResolution, 1080)
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
		m.Width, err = getIntFromMap(media, FieldNameMediaConstraintsWidth, m.Width)
		if err != nil {
			return m, err
		}
		m.Height, err = getIntFromMap(media, FieldNameMediaConstrainsHeight, m.Height)
		if err != nil {
			return m, err
		}
		return m, nil
	default:
		return MediaConstraints{}, fmt.Errorf("invalid media constraints %T: %v", media, media)
	}
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
