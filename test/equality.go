package test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/dkotik/mdsend"
)

func LettersAreEqual(a, b mdsend.Letter) func(*testing.T) {
	return func(t *testing.T) {
		if a.ID != b.ID {
			t.Log("A:", a.ID)
			t.Log("B:", b.ID)
			t.Fatal("IDs do not match")
		}
		t.Run("frontmatter is the same", func(t *testing.T) {
			if !reflect.DeepEqual(a.Frontmatter, b.Frontmatter) {
				t.Logf("A: %+v", a.Frontmatter)
				t.Logf("B: %+v", b.Frontmatter)
				t.Fatal("frontmatter does not match")
			}
		})
		if a.Content != b.Content {
			t.Log("A:", a.Content)
			t.Log("B:", b.Content)
			t.Fatal("content does not match")
		}
		if !a.CreatedAt.Truncate(time.Second).Equal(b.CreatedAt.Truncate(time.Second)) {
			t.Log("A:", a.CreatedAt.Format(time.RFC3339))
			t.Log("B:", b.CreatedAt.Format(time.RFC3339))
			t.Fatal("created at time does not match")
		}
		if !a.SentAt.Truncate(time.Second).Equal(b.SentAt.Truncate(time.Second)) {
			t.Log("A:", a.SentAt.Format(time.RFC3339))
			t.Log("B:", b.SentAt.Format(time.RFC3339))
			t.Fatal("sent at time does not match")
		}
	}
}

func AttachmentsAreEqual(a, b mdsend.Attachment) func(*testing.T) {
	return func(t *testing.T) {
		if a.Name != b.Name {
			t.Log("A:", a.Name)
			t.Log("B:", b.Name)
			t.Fatal("names do not match")
		}
		if a.LetterID != b.LetterID {
			t.Log("A:", a.LetterID)
			t.Log("B:", b.LetterID)
			t.Fatal("letter IDs do not match")
		}
		if a.ContentID != b.ContentID {
			t.Log("A:", a.ContentID)
			t.Log("B:", b.ContentID)
			t.Fatal("content IDs do not match")
		}
		if a.ContentType != b.ContentType {
			t.Log("A:", a.ContentType)
			t.Log("B:", b.ContentType)
			t.Fatal("content types do not match")
		}
		if !bytes.Equal(a.Content, b.Content) {
			t.Log("A:", a.Content)
			t.Log("B:", b.Content)
			t.Fatal("content does not match")
		}
	}
}

func DispatchesAreEqual(a, b mdsend.Dispatch) func(*testing.T) {
	return func(t *testing.T) {
		if a.ID != b.ID {
			t.Log("A:", a.ID)
			t.Log("B:", b.ID)
			t.Fatal("IDs do not match")
		}
		if a.LetterID != b.LetterID {
			t.Log("A:", a.LetterID)
			t.Log("B:", b.LetterID)
			t.Fatal("letter IDs do not match")
		}
		t.Run("senders are the same", func(t *testing.T) {
			if a.From.String() != b.From.String() {
				t.Log("A:", a.From.String())
				t.Log("B:", b.From.String())
				t.Fatal("senders do not match")
			}
		})
		t.Run("recipients are the same", func(t *testing.T) {
			if a.To.String() != b.To.String() {
				t.Log("A:", a.To.String())
				t.Log("B:", b.To.String())
				t.Fatal("recipients do not match")
			}
		})
		t.Run("headers are the same", func(t *testing.T) {
			if !reflect.DeepEqual(a.Headers, b.Headers) {
				t.Log("A:", a.Headers)
				t.Log("B:", b.Headers)
				t.Fatal("headers do not match")
			}
		})
		if a.Subject != b.Subject {
			t.Log("A:", a.Subject)
			t.Log("B:", b.Subject)
			t.Fatal("subjects do not match")
		}
		t.Run("text is the same", func(t *testing.T) {
			if a.Text != b.Text {
				t.Log("A:", a.Text)
				t.Log("B:", b.Text)
				t.Fatal("text does not match")
			}
		})
		t.Run("html is the same", func(t *testing.T) {
			if a.HTML != b.HTML {
				t.Log("A:", a.HTML)
				t.Log("B:", b.HTML)
				t.Fatal("html does not match")
			}
		})
		if !a.SentAt.Equal(b.SentAt) {
			t.Log("A:", a.SentAt.Format(time.RFC3339))
			t.Log("B:", b.SentAt.Format(time.RFC3339))
			t.Fatal("sent at time does not match")
		}
	}
}
