package test

import (
	"testing"
	"time"

	"github.com/dkotik/mdsend"
)

func MapsAreEqual(a, b map[string]any) func(*testing.T) {
	return func(t *testing.T) {
		if len(a) != len(b) {
			t.Log("A:", a)
			t.Log("B:", b)
			t.Fatal("maps do not contain the same number of elements")
		}
		for k, v1 := range a {
			if v2, ok := b[k]; !ok || v1 != v2 {
				t.Log("A:", v1)
				t.Log("B:", v2)
				t.Fatal("key value is different:", k)
			}
		}
	}
}

func LettersAreEqual(a, b mdsend.Letter) func(*testing.T) {
	return func(t *testing.T) {
		if a.ID != b.ID {
			t.Log("A:", a.ID)
			t.Log("B:", b.ID)
			t.Fatal("IDs do not match")
		}
		t.Run("frontmatter is the same", MapsAreEqual(
			a.Frontmatter,
			b.Frontmatter,
		))
		if a.Content != b.Content {
			t.Log("A:", a.Content)
			t.Log("B:", b.Content)
			t.Fatal("content does not match")
		}
		if !a.CreatedAt.Equal(b.CreatedAt) {
			t.Log("A:", a.CreatedAt.Format(time.RFC3339))
			t.Log("B:", b.CreatedAt.Format(time.RFC3339))
			t.Fatal("created at time does not match")
		}
		if !a.SentAt.Equal(b.SentAt) {
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
		// TODO: content type
		// TODO: content itself
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
		t.Run("recipients are the same", MapsAreEqual(
			a.Recipient,
			b.Recipient,
		))
		if !a.SentAt.Equal(b.SentAt) {
			t.Log("A:", a.SentAt.Format(time.RFC3339))
			t.Log("B:", b.SentAt.Format(time.RFC3339))
			t.Fatal("sent at time does not match")
		}
	}
}
