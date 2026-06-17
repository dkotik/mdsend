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
	err := NewWriter(newMockAttachmentRepository(), entropy).Write(t.Context(), b, mdsend.Message{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "Test Subject Перший",
		Text:    "text",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []mdsend.Header{
					{Name: HeaderMIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "Test Subject Перший"},
				},
			},
		),
	)
	goldie.New(t).Assert(t, "plain", b.Bytes())
}

func TestPlainMixedMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(newMockAttachmentRepository(
		mdsend.Attachment{
			Name:        "log.txt",
			Hash:        "string",
			ContentType: "text/plain; charset=utf-8",
			Content:     []byte("plain"),
		},
	), entropy).Write(t.Context(), b, mdsend.Message{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "Test Subject 3",
		Text:    "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []mdsend.Header{
					{Name: HeaderMIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "Test Subject 3"},
				},
				Children: []partDefinition{
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "text/plain; charset=utf-8"},
						},
					},
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "text/plain; charset=utf-8"},
						},
					},
				},
			},
		),
	)
	goldie.New(t).Assert(t, "plain_mixed", b.Bytes())
}

func TestAlternativeMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(newMockAttachmentRepository(), entropy).Write(t.Context(), b, mdsend.Message{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "😁 Test Subject",
		Text:    "text",
		HTML:    "<b>text</b>",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []mdsend.Header{
					{Name: HeaderMIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "😁 Test Subject"},
				},
				Children: []partDefinition{
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "text/plain; charset=utf-8"},
						},
					},
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "text/html; charset=utf-8"},
						},
					},
				},
			},
		),
	)
	goldie.New(t).Assert(t, "alternative", b.Bytes())
}

func TestMixedMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(newMockAttachmentRepository(
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
	), entropy).Write(t.Context(), b, mdsend.Message{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "😁 Test Subject",
		Text:    "text",
		HTML:    "<b>text</b>",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []mdsend.Header{
					{Name: HeaderMIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "😁 Test Subject"},
				},
				Children: []partDefinition{
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "multipart/alternative; boundary=\"HdK5vaOVwn+qpPmNt:8aYf_b0b3NcDPps:cNTLYiQokKKhDxuy_HQ+-beg2,Dlc/\""},
						},
						Children: []partDefinition{
							{
								Headers: []mdsend.Header{
									{Name: "Content-Type", Value: "text/plain; charset=utf-8"},
								},
							},
							{
								Headers: []mdsend.Header{
									{Name: "Content-Type", Value: "text/html; charset=utf-8"},
								},
							},
						},
					},
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "image/jpeg"},
						},
					},
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "image/jpeg"},
						},
					},
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "image/jpeg"},
						},
					},
				},
			},
		),
	)
	goldie.New(t).Assert(t, "mixed", b.Bytes())
}

func TestRelatedMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(newMockAttachmentRepository(
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
	), entropy).Write(t.Context(), b, mdsend.Message{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "😁 Test Subject",
		Text:    "text",
		HTML:    "<b>text</b> <img src=\"cid:string@gmail.com\" alt=\"cat\" />",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []mdsend.Header{
					{Name: HeaderMIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "😁 Test Subject"},
				},
				Children: []partDefinition{
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "multipart/alternative; boundary=\"++(aWO.mx3Zr(d'8mdBe:I+9be-=btD5/=_NgO9=_D5vOpz2Q_et=VohfwsUFilo\""},
						},
						Children: []partDefinition{
							{
								Headers: []mdsend.Header{
									{Name: "Content-Type", Value: "text/plain; charset=utf-8"},
								},
							},
							{
								Headers: []mdsend.Header{
									{Name: "Content-Type", Value: "text/html; charset=utf-8"},
								},
							},
						},
					},
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "image/jpeg"},
						},
					},
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "image/jpeg"},
						},
					},
					{
						Headers: []mdsend.Header{
							{Name: "Content-Type", Value: "image/jpeg"},
						},
					},
				},
			},
		),
	)
	goldie.New(t).Assert(t, "related", b.Bytes())
}
