package awsses

import (
	"net/mail"
	"testing"

	"github.com/dkotik/mdsend"
	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"zombiezen.com/go/sqlite"
)

func TestSendingWithAWSSES(t *testing.T) {
	// Should not require a test container, because
	// AWS offers testing facility by sending a letter
	// to [TestEmailAddress].
	if testing.Short() {
		t.Skip("SES test requires a slow test container")
	}
	t.Skip("frozen out until author can get some credentials")

	// listener, err := net.Listen("tcp", "127.0.0.1:0")
	// if err != nil {
	// 	t.Fatal("unable to obtain a local port:", err)
	// }
	// port := fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
	// if port == "" {
	// 	t.Fatal("there are no local ports available")
	// }
	// if err = listener.Close(); err != nil {
	// 	t.Fatal("unable to close local listener:", err)
	// }
	// ctx, cancel := context.WithTimeout(t.Context(), time.Second*120)
	// defer cancel()

	// mockSES, err := testcontainers.Run(
	// 	ctx,
	// 	// "domdomegg/aws-ses-v2-local", // privately hosted
	// 	"dasprid/aws-ses-v2-local",
	// 	testcontainers.WithExposedPorts(port+"/tcp"),
	// 	testcontainers.WithWaitStrategy(
	// 		wait.ForListeningPort(port+"/tcp"),
	// 		wait.ForLog("Ready to accept connections"),
	// 	),
	// )
	// testcontainers.CleanupContainer(t, mockSES)
	// if err != nil {
	// 	t.Fatal("unable to start mock SES API container:", err)
	// }

	ctx := t.Context()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		// config.WithRegion("us-east-1"),
		// config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
		// 	func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		// 		return aws.Endpoint{
		// 			URL:           "http://127.0.0.1:" + port, // Route API requests here
		// 			SigningRegion: "us-east-1",
		// 		}, nil
		// 	})),
	)
	if err != nil {
		// t.Skip("no credentials could be obtained:", err)
		t.Fatal("unable to set up AWS SES configuration:", err)
	}

	conn, err := sqlite.OpenConn(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	q, err := sqliteQ.New(conn, "")
	if err != nil {
		t.Fatal(err)
	}
	sesMailer, err := New(q, sesv2.NewFromConfig(cfg))
	if err != nil {
		t.Fatal("unable to setup an AWS SES mailer:", err)
	}

	messageID, err := sesMailer.SendMail(ctx, mdsend.Message{
		From: mail.Address{
			Name:    "Test Sender",
			Address: "test@test.com",
		},
		To: mail.Address{
			Name:    "Test Recipient",
			Address: TestEmailAddress,
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
}
