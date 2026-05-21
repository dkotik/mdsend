package mime

import (
	"bytes"
	"math/rand/v2"
	"net/textproto"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func TestBareMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	if err := (Message{
		Entropy:     rand.NewChaCha8([32]byte{}),
		Header:      textproto.MIMEHeader{},
		Text:        "text",
		HTML:        "<b>text</b>",
		Attachments: []Attachment{},
	}).Encode(b); err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "textonly", b.Bytes())
}

func TestMessageEncodingWithAttachments(t *testing.T) {
	b := &bytes.Buffer{}
	if err := (Message{
		Entropy:     rand.NewChaCha8([32]byte{}),
		Header:      textproto.MIMEHeader{},
		Text:        "text",
		HTML:        "<b>text</b>",
		Attachments: []Attachment{},
	}).Encode(b); err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "attachments", b.Bytes())
}

func TestMessageEncodingWithEmbeddedAttachments(t *testing.T) {
	b := &bytes.Buffer{}
	if err := (Message{
		Entropy:     rand.NewChaCha8([32]byte{}),
		Header:      textproto.MIMEHeader{},
		Text:        "text",
		HTML:        "<b>text</b>",
		Attachments: []Attachment{},
	}).Encode(b); err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "embedded", b.Bytes())
}
