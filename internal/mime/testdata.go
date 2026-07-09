package mime

import _ "embed" // for images

//go:embed testdata/image/cat.jpg
var Cat []byte

//go:embed testdata/image/panda.jpg
var Panda []byte

//go:embed testdata/image/chamillion.jpg
var Chamillion []byte
