package queue

import (
	"os"
	"testing"
)

func TestAttachmentData(t *testing.T) {
	data, err := os.ReadFile("../internal/testdata/pass/a.md")
	if err != nil {
		t.Fatal(err)
	}
	attachmentData := NewAttachmentData(data)
	t.Log(attachmentData.DataEncodedToBase64)
}
