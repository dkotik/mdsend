package mdsend

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"path"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/cespare/xxhash/v2"
	"github.com/dkotik/mdsend/internal/media"
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

type Attachment struct {
	LetterID string
	Name     string
	Source   string
	// Hash                     uint64 // for XXHash2
	Hash string
	// mimeEncodedBase64Content []byte

	// ContentID is the ID of the inline attachment to use in the message.
	// It must conform to RFC 2392 format, including the angle brackets:
	// <hash@domain.com>. Leave empty for standalone attachments that
	// are not referenced by other parts of the message.
	ContentID   string
	ContentType string
	Content     []byte
}

func NewAttachment(b []byte, constraints media.Constraints) (a Attachment, err error) {
	a.Content, a.ContentType, err = constraints.ApplyTo(b)
	if err != nil {
		return a, err
	}
	a.Hash = media.DeterministicHashStringOf(a.Content)
	return a, nil
}

func NewAttachmentFromFile(fs fs.FS, p string, constraints media.Constraints) (a Attachment, err error) {
	file, err := fs.Open(p)
	if err != nil {
		return a, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	b, err := io.ReadAll(file)
	if err != nil {
		return a, err
	}
	a, err = NewAttachment(b, constraints)
	if err != nil {
		return a, err
	}
	a.Name = path.Base(p)
	return a, err
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
