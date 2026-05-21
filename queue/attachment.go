package queue

import (
	"bytes"
	"io"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/cespare/xxhash/v2"
	"github.com/dkotik/mdsend/internal/mime"
)

type Attachment struct {
	ID          string
	MessageID   string
	Name        string
	Title       string
	Path        string
	ContentType string
}

type AttachmentData struct {
	ID                  string
	AttachmentID        string
	DataEncodedToBase64 string
}

func NewAttachmentData(data []byte) AttachmentData {
	b := &bytes.Buffer{}
	// w := base64.StdEncoding.EncodeToString(src []byte)
	hash := xxhash.New()

	var err error
	enc := mime.NewEncoderBase64(b)
	if _, err = io.Copy(
		io.MultiWriter(hash, enc),
		bytes.NewReader(data),
	); err != nil {
		panic(err)
	}
	if err = enc.Close(); err != nil {
		panic(err)
	}
	return AttachmentData{
		ID:                  base58.Encode(hash.Sum(nil)),
		DataEncodedToBase64: b.String(),
	}
}
