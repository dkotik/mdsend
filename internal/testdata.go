package internal

import _ "embed" // for images

//go:embed testdata/cat.jpg
var Cat []byte

//go:embed testdata/panda.jpg
var Panda []byte

//go:embed testdata/chamillion.jpg
var Chamillion []byte
