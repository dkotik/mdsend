package mdsend

import (
	"errors"
	"fmt"
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
