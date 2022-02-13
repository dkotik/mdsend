package mdsend

import "io"

type Attachment struct {
	Hash                     [8]byte // for XXHash2
	mimeEncodedBase64Content []byte
}

func (a *Attachment) WriteTo(w io.Writer) (int64, error) {

	return 0, nil
}
