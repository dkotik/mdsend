package mdsend

import (
	"path/filepath"
	"testing"

	"github.com/dkotik/mdsend/media"
)

func TestAttachments(t *testing.T) {
	loader, err := New(media.NewUnsafeUnconstrainedFileSystem(), Defaults{})
	if err != nil {
		t.Fatal(err)
	}
	_, attachments, err := loader.LoadLetter(
		t.Context(),
		filepath.Join("testdata", "attachments", "letter.md"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(attachments) != 4 {
		t.Errorf("expected 4 attachments, got %d", len(attachments))
	}

	tcs := []struct {
		Name        string
		ContentID   string
		ContentType string
	}{
		{Name: "cat.jpg", ContentID: "", ContentType: "image/jpeg"},
		{Name: "attachment.txt", ContentID: "", ContentType: "text/plain; charset=utf-8"},
		{Name: "chamillion.jpg", ContentID: "hydNeKLUx3Y@test.com", ContentType: "image/jpeg"},
		{Name: "panda.jpg", ContentID: "S2EnMRZNrA7@test.com", ContentType: "image/jpeg"},
	}

	for i, attachment := range attachments {
		if i > len(tcs)-1 {
			t.Error("extra attachment", i+1, attachment.Name)
			continue
		}
		t.Log("Attached:", attachment.Name)
		if attachment.Name != tcs[i].Name {
			t.Errorf("expected attachment name %s, got %s", tcs[i].Name, attachment.Name)
		}
		if attachment.ContentID != tcs[i].ContentID {
			t.Errorf("expected attachment content ID %q, got %q", tcs[i].ContentID, attachment.ContentID)
		}
		if attachment.ContentType != tcs[i].ContentType {
			t.Errorf("expected attachment content type %s, got %s", tcs[i].ContentType, attachment.ContentType)
		}
	}
}
