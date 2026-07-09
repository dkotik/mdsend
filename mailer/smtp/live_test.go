package smtp

import (
	"net/mail"
	"os"
	"strings"
	"testing"

	"github.com/dkotik/mdsend"
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
	constraints := media.Constraints{
		Width:   100,
		Height:  100,
		Quality: 20,
	}

	cat, err := mdsend.NewAttachment(mime.Cat, constraints)
	cat.Name = "cat.jpg"
	cat.LetterID = testLetterID
	if err = config.Queue.CreateAttachment(ctx, cat); err != nil {
		t.Fatal(err)
	}
	chamillion, err := mdsend.NewAttachment(mime.Chamillion, constraints)
	chamillion.Name = "chamillion.jpg"
	chamillion.LetterID = testLetterID
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
		HTML:    "<html><body><h1>test</h1><p>test paragraph</p><p>test paragraph 2</p><p><img src=\"cid:" + cat.Hash + "@testdomain.com\" alt=\"cat\" /></p></body></html>",
	})

	if err != nil {
		t.Fatal(err)
	}

	if messageID == "" {
		t.Fatal("message ID is empty")
	}
	t.Log("message ID:", messageID)
}
