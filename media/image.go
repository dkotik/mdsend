package media

import (
	"image"

	"github.com/nfnt/resize"
)

func Resize(image image.Image, width, height uint) image.Image {
	return resize.Thumbnail(
		width,
		height,
		image,
		resize.Lanczos3,
	)
}
