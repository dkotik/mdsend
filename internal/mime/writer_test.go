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
							{Name: "Content-Type", Value: "multipart/alternative; boundary=\"6H48lprQ5jsIY3Q8:6x9i/'O6SWNdslloeYLjBi)Riw3.d6lAvlNNTU9Zg8_:ZaT\""},
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
		/*
			multipart/mixed
			|- multipart/alternative
			|  |- text/plain
			|  `- multipart/related
			|     |- text/html
			|     `- inline_attachments...
			`- attachments...
		*/
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
							{Name: "Content-Type", Value: "multipart/alternative; boundary=\"/y-Da+X'9'De5C,n(2Sv(wD1heyVBka:XtW:+5,e+(:.e2Vj=gNyi'RDve9)hyla\""},
						},
						Children: []partDefinition{
							{
								Headers: []mdsend.Header{
									{Name: "Content-Type", Value: "text/plain; charset=utf-8"},
								},
							},
							{
								Headers: []mdsend.Header{
									{Name: "Content-Type", Value: "multipart/related; boundary=\"y=BbvMTZKArDv14o2eR7/y'?ok:fkQaf=qstd+=gwzmkK:zAMh70OtiBfDfIgFyF\""},
								},
								Children: []partDefinition{
									{
										Headers: []mdsend.Header{
											{Name: "Content-Type", Value: "text/html; charset=utf-8"},
										},
									},
									{
										Headers: []mdsend.Header{
											{Name: "Content-Type", Value: "image/jpeg"},
										},
									},
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
				},
			},
		),
	)
	goldie.New(t).Assert(t, "related", b.Bytes())
}
