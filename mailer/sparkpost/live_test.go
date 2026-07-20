package sparkpost

import (
	"errors"
	"net/mail"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal/media"
)

func TestLiveSend(t *testing.T) {
	destination := strings.TrimSpace(os.Getenv(EnvironmentEmailTo))
	if destination == "" {
		t.Skip("environment variable " + EnvironmentEmailTo + " is not set")
	}

	config := getLiveConfigOrSkip(t)
	mg, err := New(config)
	if err != nil {
		if errors.Is(err, ErrMissingAPIKey) {
			t.Skip("missing API key")
		}
		t.Fatal(err)
	}

	ctx := t.Context()
	constraints := media.Constraints{
		Width:   100,
		Height:  100,
		Quality: 20,
	}

	cat, err := mdsend.NewAttachment(media.Cat, constraints)
	cat.Name = "cat.jpg"
	cat.LetterID = testLetterID
	if err = config.Queue.CreateAttachment(ctx, cat); err != nil {
		t.Fatal(err)
	}
	chamillion, err := mdsend.NewAttachment(media.Chamillion, constraints)
	chamillion.Name = "chamillion.jpg"
	chamillion.LetterID = testLetterID
	if err = config.Queue.CreateAttachment(ctx, chamillion); err != nil {
		t.Fatal(err)
	}

	messageID, err := mg.SendMail(ctx, mdsend.Message{
		SeedKey:  time.Now().Truncate(time.Minute).Format(time.RFC3339),
		LetterID: testLetterID,
		From: mail.Address{
			Name:    "Test Sender",
			Address: "test@test.com",
		},
		To: mail.Address{
			Name:    "Test Recipient",
			Address: destination,
		},
		Subject: "live resent send test",
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
