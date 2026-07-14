package media

import (
	_ "embed" // for images
)

var (
	//go:embed testdata/cat.jpg
	Cat []byte
	//go:embed testdata/panda.jpg
	Panda []byte
	//go:embed testdata/chamillion.jpg
	Chamillion []byte
)
