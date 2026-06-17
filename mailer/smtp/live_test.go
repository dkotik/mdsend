package smtp

import (
	"net/mail"
	"os"
	"strings"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal"
	"github.com/dkotik/mdsend/internal/media"
	"github.com/dkotik/mdsend/internal/mime"
)

const testLetterID = "test-letter-id"

func TestSend(t *testing.T) {
	destination := strings.TrimSpace(os.Getenv(EnvironmentTestTo))
	if destination == "" {
		t.Skip("environment variable " + EnvironmentTestTo + " is not set")
	}

	config := getLiveConfigOrSkip(t)
	sender, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	constraints := mdsend.MediaConstraints{
		Width:   100,
		Height:  100,
		Quality: 20,
	}
	cat, err := media.Compress(mdsend.Attachment{
		LetterID:    testLetterID,
		Content:     internal.Cat,
		Name:        "cat.jpg",
		ContentType: mime.ContentTypeImageJPEG,
		Hash:        "cat",
	}, constraints)
	if err = config.Queue.CreateAttachment(ctx, cat); err != nil {
		t.Fatal(err)
	}
	chamillion, err := media.Compress(mdsend.Attachment{
		LetterID:    testLetterID,
		Content:     internal.Chamillion,
		Name:        "chamillion.jpg",
		ContentType: mime.ContentTypeImageJPEG,
		Hash:        "chamillion",
	}, constraints)
	if err = config.Queue.CreateAttachment(ctx, chamillion); err != nil {
		t.Fatal(err)
	}

	messageID, err := sender.SendMail(ctx, mdsend.Message{
		ID:       "testMessage",
		LetterID: testLetterID,
		From: mail.Address{
			Name:    "Test Sender",
			Address: "test@test.com",
		},
		To: mail.Address{
			Name:    "Test Recipient",
			Address: destination,
		},
		Subject: "live SMTP send test",
		Text:    "test text",
		HTML:    "<html><body><h1>test</h1><p>test paragraph</p><p>test paragraph 2</p><p><img src=\"cid:cat@testdomain.com\" alt=\"cat\" /></p></body></html>",
	})

	if err != nil {
		t.Fatal(err)
	}

	if messageID == "" {
		t.Fatal("message ID is empty")
	}
	t.Log("message ID:", messageID)
}
