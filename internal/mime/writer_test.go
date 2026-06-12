package mime

import (
	"bytes"
	"math/rand/v2"
	"net/mail"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal"
	"github.com/sebdah/goldie/v2"
)

var entropy = rand.New(rand.NewPCG(0, 0))

func TestPlainMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(b, newMockAttachmentRepository(), entropy).Write(t.Context(), mdsend.Dispatch{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "Test Subject",
		Text:    "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "plain", b.Bytes())
}

func TestPlainMixedMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(b, newMockAttachmentRepository(
		mdsend.Attachment{
			Name:        "log.txt",
			Hash:        "string",
			ContentType: "text/plain",
			Content:     []byte("plain"),
		},
	), entropy).Write(t.Context(), mdsend.Dispatch{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "Test Subject",
		Text:    "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "plain_mixed", b.Bytes())
}

func TestAlternativeMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(b, newMockAttachmentRepository(), entropy).Write(t.Context(), mdsend.Dispatch{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "😁 Test Subject",
		Text:    "text",
		HTML:    "<b>text</b>",
	})
	if err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "alternative", b.Bytes())
}

func TestMixedMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(b, newMockAttachmentRepository(
		mdsend.Attachment{
			Name:        "cat.jpg",
			Hash:        "string",
			ContentType: ContentTypeImageJPEG,
			Content:     internal.Cat,
		},
		mdsend.Attachment{
			Name:        "panda.jpg",
			Hash:        "string1",
			ContentType: ContentTypeImageJPEG,
			Content:     internal.Panda,
		},
		mdsend.Attachment{
			Name:        "chamillion.jpg",
			Hash:        "string2",
			ContentType: ContentTypeImageJPEG,
			Content:     internal.Chamillion,
		},
	), entropy).Write(t.Context(), mdsend.Dispatch{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "😁 Test Subject",
		Text:    "text",
		HTML:    "<b>text</b>",
	})
	if err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "mixed", b.Bytes())
}

func TestRelatedMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(b, newMockAttachmentRepository(
		mdsend.Attachment{
			Name:        "cat.jpg",
			Hash:        "string",
			ContentType: ContentTypeImageJPEG,
			Content:     internal.Cat,
		},
		mdsend.Attachment{
			Name:        "panda.jpg",
			Hash:        "string1",
			ContentType: ContentTypeImageJPEG,
			Content:     internal.Panda,
		},
		mdsend.Attachment{
			Name:        "chamillion.jpg",
			Hash:        "string2",
			ContentType: ContentTypeImageJPEG,
			Content:     internal.Chamillion,
		},
	), entropy).Write(t.Context(), mdsend.Dispatch{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "😁 Test Subject",
		Text:    "text",
		HTML:    "<b>text</b> <img src=\"cid:string@gmail.com\" alt=\"cat\" />",
	})
	if err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "related", b.Bytes())
}

/*
func TestMessageEncodingWithAttachments(t *testing.T) {
	b := &bytes.Buffer{}
	if err := (Message{
		Entropy: rand.NewChaCha8([32]byte{}),
		Header:  textproto.MIMEHeader{},
		Text:    "text",
		HTML:    "<b>text</b>",
		Attachments: []Attachment{
			{
				Name:        "cat.jpg",
				ContentType: ContentTypeImageJPEG,
				Content:     internal.Cat,
			},
			{
				Name:        "panda.jpg",
				ContentType: ContentTypeImageJPEG,
				Content:     internal.Panda,
			},
			{
				Name:        "chamillion.jpg",
				ContentType: ContentTypeImageJPEG,
				Content:     internal.Chamillion,
			},
		},
	}).Encode(b); err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "attachments", b.Bytes())
}

func TestMessageEncodingWithEmbeddedAttachments(t *testing.T) {
	b := &bytes.Buffer{}
	if err := (Message{
		Entropy: rand.NewChaCha8([32]byte{}),
		Header:  textproto.MIMEHeader{},
		Text:    "text",
		HTML:    "<b>text</b> <img src=\"cid:cat\" alt=\"cat\" />",
		Attachments: []Attachment{
			{
				Name:        "panda.jpg",
				ContentType: ContentTypeImageJPEG,
				Content:     internal.Panda,
			},
			{
				Name:        "chamillion.jpg",
				ContentType: ContentTypeImageJPEG,
				Content:     internal.Chamillion,
			},
		},
		EmbeddedAttachments: []EmbeddedAttachment{
			{
				Name:        "cat.jpg",
				ContentType: ContentTypeImageJPEG,
				Content:     internal.Cat,
			},
		},
	}).Encode(b); err != nil {
		t.Fatal(err)
	}
	goldie.New(t).Assert(t, "embedded", b.Bytes())
}
*/
