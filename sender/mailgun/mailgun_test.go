package mailgun

import (
	"errors"
	"net/mail"
	"testing"

	"github.com/dkotik/mdsend"
)

func TestMailgunSender(t *testing.T) {
	mg, err := New(Configuration{
		TestMode: true,
	})
	if err != nil {
		if errors.Is(err, ErrMissingAPIKey) {
			t.Skip("missing API key")
		}
		if errors.Is(err, ErrMissingDomain) {
			t.Skip("missing API domain")
		}
		t.Fatal(err)
	}

	messageID, err := mg.Send(t.Context(), mdsend.Dispatch{
		From: mail.Address{
			Name:    "Test Sender",
			Address: "test@test.com",
		},
		To: mail.Address{
			Name:    "Test Recipient",
			Address: "recipient@test.com",
		},
		Subject: "test subject",
		Text:    "test text",
	})

	if err != nil {
		t.Fatal(err)
	}

	if messageID == "" {
		t.Fatal("message ID is empty")
	}
	t.Log("message ID:", messageID)
	// t.Fail()
}
