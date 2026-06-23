package mdsend

import (
	"errors"
	"fmt"

	"github.com/dkotik/mdsend/internal/media"
)

func (l Letter) GetMediaConstraints() (m media.Constraints, err error) {
	switch media := l.Frontmatter[FieldNameMediaContraints].(type) {
	case nil:
		return m, nil
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
		return m, fmt.Errorf("invalid media constraints %T: %v", media, media)
	}
}
