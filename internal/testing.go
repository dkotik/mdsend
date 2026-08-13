package internal

import (
	_ "embed" // for images
)

var (
	//go:embed testdata/image/cat.jpg
	Cat []byte
	//go:embed testdata/image/panda.jpg
	Panda []byte
	//go:embed testdata/image/chamillion.jpg
	Chamillion []byte
)
