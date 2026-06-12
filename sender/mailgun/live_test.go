package mailgun

import (
	"errors"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal"
	"github.com/dkotik/mdsend/internal/mime"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"zombiezen.com/go/sqlite"
)

func TestLiveSend(t *testing.T) {
	destination := strings.TrimSpace(os.Getenv(EnvironmentEmailTo))
	if destination == "" {
		t.Skip("environment variable " + EnvironmentEmailTo + " is not set")
	}
	liveTestLockFile := filepath.Join("testdata", "live_test.lock")
	_, err := os.Stat(liveTestLockFile)
	if err == nil {
		t.Skip("live test lock file exists")
	}
	if !os.IsNotExist(err) {
		t.Fatal("failed to check live test lock file: ", err)
	}
	if err = os.WriteFile(liveTestLockFile, []byte("delete this file to fire off live test"), 0600); err != nil {
		t.Fatal("failed to create live test lock file: ", err)
	}

	conn, err := sqlite.OpenConn(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err = conn.Close(); err != nil {
			t.Fatal(err)
		}
	})
	q, err := sqliteQ.New(conn, "")
	if err != nil {
		t.Fatal(err)
	}
	mg, err := New(Configuration{
		Queue:    q,
		TestMode: false,
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

	ctx := t.Context()

	err = q.CreateAttachment(ctx, mdsend.Attachment{
		LetterID:    testLetterID,
		Content:     internal.Cat,
		Name:        "cat.jpg",
		ContentType: mime.ContentTypeImageJPEG,
		Hash:        "cat",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = q.CreateAttachment(ctx, mdsend.Attachment{
		LetterID:    testLetterID,
		Content:     internal.Chamillion,
		Name:        "chamillion.jpg",
		ContentType: mime.ContentTypeImageJPEG,
		Hash:        "chamillion",
	})
	if err != nil {
		t.Fatal(err)
	}

	messageID, err := mg.Send(ctx, mdsend.Dispatch{
		LetterID: testLetterID,
		From: mail.Address{
			Name:    "Test Sender",
			Address: "test@test.com",
		},
		To: mail.Address{
			Name:    "Test Recipient",
			Address: destination,
		},
		Subject: "live Mailgun send test",
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
