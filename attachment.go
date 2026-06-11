package mdsend

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/cespare/xxhash/v2"
)

type AttachmentError uint8

const (
	ErrInvalidAttachment AttachmentError = iota
	ErrDuplicateAttachment
)

func (err AttachmentError) Error() string {
	switch err {
	case ErrInvalidAttachment:
		return "invalid attachment"
	case ErrDuplicateAttachment:
		return "duplicate attachment"
	default:
		return ""
	}
}

func NewAttachment(p string) (*Attachment, error) {
	r, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	contentType, b, err := sniffLoadCompressAttachment(r)
	if err != nil {
		return nil, err
	}

	h := xxhash.New()
	_, err = io.Copy(h, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	id := h.Sum64()
	if attachment, ok := attachments[id]; ok {
		return attachment, nil
	}

	header := textproto.MIMEHeader{}
	header.Set(`Content-Transfer-Encoding`, `base64`)
	header.Set(`Content-Type`, contentType)
	fname := filepath.Base(p)
	header.Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename=%q`, fname))
	// if inline {
	//     header.Set(`Content-Disposition`, fmt.Sprintf(`inline; filename=%q`, fname))
	//     header.Set(`Content-ID`, fmt.Sprintf(`<%s>`, fname))
	// } else {
	// }

	var final bytes.Buffer
	if err = MIMEHeaderTo(&final, header); err != nil {
		return nil, err
	}
	if err = MIMEBase64To(&final, bytes.NewReader(b)); err != nil {
		return nil, err
	}

	return &Attachment{
		Name:                     fname,
		Source:                   p,
		Hash:                     base58.Encode(h.Sum(nil)),
		mimeEncodedBase64Content: final.Bytes(),
	}, nil
}

// Looks at the first bytes to assertain the MIME content type, rewinds to the beginning, copies the rest of the source into memory, compresses content, if there is a matching compressor. Returns MIME content type and the buffer containing the content.
func sniffLoadCompressAttachment(source io.ReadSeeker) (string, []byte, error) {
	var b, sniff bytes.Buffer
	_, err := io.CopyN(&sniff, source, 512)
	if err != nil {
		return "", nil, err
	}
	if _, err = source.Seek(0, io.SeekStart); err != nil {
		return "", nil, err
	}
	if _, err = io.Copy(&b, source); err != nil { // TODO: add max file size?
		return "", nil, err
	}

	contentType := http.DetectContentType(sniff.Bytes())
	if compressor, ok := compressorsByContentType[contentType]; ok {
		resized, err := compressor(&b)
		if err != nil {
			return "", nil, err
		}
		return contentType, resized.Bytes(), nil
	}
	return contentType, b.Bytes(), nil
}

type Attachment struct {
	LetterID string
	Name     string
	Source   string
	// Hash                     uint64 // for XXHash2
	Hash                     string
	mimeEncodedBase64Content []byte

	ContentType string
	Content     []byte
}

func (a Attachment) WithUpdatedHash() Attachment {
	hash := xxhash.New()
	var err error
	if _, err = io.Copy(
		hash,
		bytes.NewReader(a.Content),
	); err != nil {
		panic(err)
	}
	a.Hash = base58.Encode(hash.Sum(nil))
	// a.Hash = hash.Sum64()
	return a
}

// deprecate
func (a *Attachment) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, bytes.NewReader(a.mimeEncodedBase64Content))
}

var attachments = make(map[uint64]*Attachment)
