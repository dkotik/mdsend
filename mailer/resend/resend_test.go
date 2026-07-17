package resend

import (
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dkotik/mdsend"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"zombiezen.com/go/sqlite"
)

const testLetterID = "test-letter-id"

var getLiveConfigOrSkip func(t *testing.T) Configuration

func TestMain(m *testing.M) {
	conn, err := sqlite.OpenConn(":memory:")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = conn.Close(); err != nil {
			panic(err)
		}
	}()
	q, err := sqliteQ.New(conn, "")
	if err != nil {
		panic(err)
	}

	liveConfig := Configuration{
		Queue:    q,
		APIKey:   strings.TrimSpace(os.Getenv(EnvironmentKey)),
		TestMode: false,
	}
	getLiveConfigOrSkip = func() func(*testing.T) Configuration {
		if liveConfig.APIKey == "" {
			return func(t *testing.T) Configuration {
				t.Skip("no live test API environment key set: " + EnvironmentKey)
				return liveConfig
			}
		}

		liveTestLockFile := filepath.Join("testdata", "live_test.lock")
		_, err := os.Stat(liveTestLockFile)
		if err == nil {
			return func(t *testing.T) (s Configuration) {
				t.Skip("live test lock file exists: remove it to allow live tests")
				return liveConfig
			}
		}
		if !os.IsNotExist(err) {
			panic(fmt.Errorf("failed to check live test lock file: %w", err))
		}
		if err = os.WriteFile(liveTestLockFile, []byte("delete this file to fire off live test"), 0600); err != nil {
			panic(fmt.Errorf("failed to create live test lock file: %w", err))
		}
		return func(t *testing.T) (s Configuration) {
			return liveConfig
		}
	}()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestSend(t *testing.T) {
	t.Log("note that Resend will not send mail from unverified domain")
	config := getLiveConfigOrSkip(t)
	// config.TestMode = true
	mg, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	messageID, err := mg.SendMail(t.Context(), mdsend.Message{
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
