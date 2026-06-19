package loader

import (
	"os"
	"testing"

	"github.com/dkotik/mdsend"
)

func TestRecipientList(t *testing.T) {
	expected := []string{
		"first@testmail.yaml",
		"second@testmail.yaml",
		"third@testmail.yaml",
		"fourth@testmail.yaml",
		"first@testmail.json",
		"second@testmail.json",
		"third@testmail.json",
		"fourth@testmail.json",
		"first@testmail.cue",
		"second@testmail.cue",
		"third@testmail.cue",
		"fourth@testmail.cue",
		"first@testmail.toml",
		"second@testmail.toml",
		"third@testmail.toml",
		"fourth@testmail.toml",
		"bcc@test.com",
	}
	cursor := 0
	lastCursor := len(expected)

	// fs, err := fs.Sub(os.DirFS("testdata"), "testdata")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if f, err := os.DirFS("testdata").Open("recipients.yaml"); err != nil {
	// 	t.Fatal(err)
	// } else {
	// 	f.Close()
	// }
	l := loader{
		FS: os.DirFS("testdata"),
		// Cache: make(map[string][]byte),
	}
	for recipient, err := range l.eachRecipient(map[string]any{
		"to": []any{
			// "first@testmail.com",
			"./recipients.yaml",
			"./recipients.cue",
			// "recipients.yaml",
			// "recipients.yaml",
			"bcc@test.com",
		},
	}, ".") {
		if err != nil {
			t.Fatal(err)
		}
		t.Log(recipient)
		email, ok := recipient[mdsend.FieldNameEmail]
		if !ok {
			t.Fatal("recipient does not contain an email address")
		}
		if expected[cursor] != email {
			t.Fatal("unexpected email:", email, "vs", expected[cursor])
		}
		cursor++
		// if cursor == lastCursor {
		// 	t.Fatal("too many contacts loaded")
		// }
	}
	if cursor != lastCursor {
		t.Fatal("too few contacts loaded:", cursor, "vs", lastCursor)
	}
}
