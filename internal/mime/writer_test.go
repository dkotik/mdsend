package mime

import (
	"bytes"
	"math/rand/v2"
	"net/mail"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/header"
	"github.com/dkotik/mdsend/internal/media"
	"github.com/sebdah/goldie/v2"
)

var entropy = rand.New(rand.NewPCG(0, 0))

const longText = "Ut non ipsum orci. Sed vel sollicitudin lectus. Sed quis arcu laoreet, euismod enim ac, gravida justo. Vivamus sed nunc nec orci dignissim consequat aliquam consequat massa. Sed et porta leo. Pellentesque gravida enim vel dui vulputate, quis vulputate turpis gravida. Integer lacinia, elit at venenatis posuere, odio turpis fringilla ipsum, eget tempor magna lacus sit amet nunc. Vivamus elementum massa non sapien luctus ornare. Etiam mattis est eu diam porttitor suscipit. Phasellus tortor tellus, porta sit amet sagittis nec, rhoncus tincidunt sem. Cras porttitor pharetra turpis sit amet viverra. Nulla egestas leo metus, quis tincidunt dui varius non. Fusce vel eros leo. Suspendisse iaculis pulvinar elementum. In nec dapibus lacus. Curabitur viverra, nisi sit amet sagittis aliquet, ligula quam consectetur enim, et malesuada sapien nulla sed metus. Ut non ipsum orci. Sed vel sollicitudin lectus. Sed quis arcu laoreet, euismod enim ac, gravida justo. Vivamus sed nunc nec orci dignissim consequat aliquam consequat massa. Sed et porta leo. Pellentesque gravida enim vel dui vulputate, quis vulputate turpis gravida. Integer lacinia, elit at venenatis posuere, odio turpis fringilla ipsum, eget tempor magna lacus sit amet nunc. Vivamus elementum massa non sapien luctus ornare. Etiam mattis est eu diam porttitor suscipit. Phasellus tortor tellus, porta sit amet sagittis nec, rhoncus tincidunt sem. Cras porttitor pharetra turpis sit amet viverra. Nulla egestas leo metus, quis tincidunt dui varius non. Fusce vel eros leo. Suspendisse iaculis pulvinar elementum. In nec dapibus lacus. Curabitur viverra, nisi sit amet sagittis aliquet, ligula quam consectetur enim, et malesuada sapien nulla sed metus. Ut non ipsum orci. Sed vel sollicitudin lectus. Sed quis arcu laoreet, euismod enim ac, gravida justo. Vivamus sed nunc nec orci dignissim consequat aliquam consequat massa. Sed et porta leo. Pellentesque gravida enim vel dui vulputate, quis vulputate turpis gravida. Integer lacinia, elit at venenatis posuere, odio turpis fringilla ipsum, eget tempor magna lacus sit amet nunc. Vivamus elementum massa non sapien luctus ornare. Etiam mattis est eu diam porttitor suscipit. Phasellus tortor tellus, porta sit amet sagittis nec, rhoncus tincidunt sem. Cras porttitor pharetra turpis sit amet viverra. Nulla egestas leo metus, quis tincidunt dui varius non. Fusce vel eros leo. Suspendisse iaculis pulvinar elementum. In nec dapibus lacus. Curabitur viverra, nisi sit amet sagittis aliquet, ligula quam consectetur enim, et malesuada sapien nulla sed metus."

func TestPlainMessageEncoding(t *testing.T) {
	b := &bytes.Buffer{}
	err := NewWriter(newMockAttachmentRepository(), entropy).Write(t.Context(), b, mdsend.Message{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "Test Subject Перший",
		Text:    longText,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []header.Header{
					{Name: header.MIMEVersion, Value: "1.0"},
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
		Text:    longText,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []header.Header{
					{Name: header.MIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "Test Subject 3"},
				},
				Children: []partDefinition{
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "text/plain; charset=utf-8"},
						},
					},
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "text/plain; charset=utf-8"},
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
		Text:    longText,
		HTML:    "<b>" + longText + "</b>",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []header.Header{
					{Name: header.MIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "😁 Test Subject"},
				},
				Children: []partDefinition{
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "text/plain; charset=utf-8"},
						},
					},
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "text/html; charset=utf-8"},
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
			Content:     media.Cat,
		},
		mdsend.Attachment{
			Name:        "panda.jpg",
			Hash:        "string1",
			ContentType: ContentTypeImageJPEG,
			Content:     media.Panda,
		},
		mdsend.Attachment{
			Name:        "chamillion.jpg",
			Hash:        "string2",
			ContentType: ContentTypeImageJPEG,
			Content:     media.Chamillion,
		},
	), entropy).Write(t.Context(), b, mdsend.Message{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "😁 Test Subject",
		Text:    longText,
		HTML:    "<b>" + longText + "</b>",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"validate structure",
		ValidateMessageStructure(
			bytes.NewReader(b.Bytes()),
			partDefinition{
				Headers: []header.Header{
					{Name: header.MIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "😁 Test Subject"},
				},
				Children: []partDefinition{
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "multipart/alternative; boundary=\"6H48lprQ5jsIY3Q8:6x9i/'O6SWNdslloeYLjBi)Riw3.d6lAvlNNTU9Zg8_:ZaT\""},
						},
						Children: []partDefinition{
							{
								Headers: []header.Header{
									{Name: header.ContentType, Value: "text/plain; charset=utf-8"},
								},
							},
							{
								Headers: []header.Header{
									{Name: header.ContentType, Value: "text/html; charset=utf-8"},
								},
							},
						},
					},
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "image/jpeg"},
						},
					},
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "image/jpeg"},
						},
					},
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "image/jpeg"},
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
			Content:     media.Cat,
		},
		mdsend.Attachment{
			Name:        "panda.jpg",
			Hash:        "string1",
			ContentType: ContentTypeImageJPEG,
			Content:     media.Panda,
		},
		mdsend.Attachment{
			Name:        "chamillion.jpg",
			Hash:        "string2",
			ContentType: ContentTypeImageJPEG,
			Content:     media.Chamillion,
		},
	), entropy).Write(t.Context(), b, mdsend.Message{
		From:    mail.Address{Name: "Sender", Address: "sender@example.com"},
		To:      mail.Address{Name: "Recipient", Address: "recipient@example.com"},
		Subject: "😁 Test Subject",
		Text:    longText,
		HTML:    "<b>" + longText + "</b> <img src=\"cid:string@gmail.com\" alt=\"cat\" />",
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
				Headers: []header.Header{
					{Name: header.MIMEVersion, Value: "1.0"},
					{Name: "From", Value: "\"Sender\" <sender@example.com>"},
					{Name: "To", Value: "\"Recipient\" <recipient@example.com>"},
					{Name: "Subject", Value: "😁 Test Subject"},
				},
				Children: []partDefinition{
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "multipart/alternative; boundary=\"/y-Da+X'9'De5C,n(2Sv(wD1heyVBka:XtW:+5,e+(:.e2Vj=gNyi'RDve9)hyla\""},
						},
						Children: []partDefinition{
							{
								Headers: []header.Header{
									{Name: header.ContentType, Value: "text/plain; charset=utf-8"},
								},
							},
							{
								Headers: []header.Header{
									{Name: header.ContentType, Value: "multipart/related; boundary=\"y=BbvMTZKArDv14o2eR7/y'?ok:fkQaf=qstd+=gwzmkK:zAMh70OtiBfDfIgFyF\""},
								},
								Children: []partDefinition{
									{
										Headers: []header.Header{
											{Name: header.ContentType, Value: "text/html; charset=utf-8"},
										},
									},
									{
										Headers: []header.Header{
											{Name: header.ContentType, Value: "image/jpeg"},
										},
									},
								},
							},
						},
					},
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "image/jpeg"},
						},
					},
					{
						Headers: []header.Header{
							{Name: header.ContentType, Value: "image/jpeg"},
						},
					},
				},
			},
		),
	)
	goldie.New(t).Assert(t, "related", b.Bytes())
}
