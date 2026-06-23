package mdsend

import (
	"os"
	"testing"
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
		"first@testmail.yaml",
		"second@testmail.yaml",
		"third@testmail.yaml",
		"fourth@testmail.yaml",
		"first@testmail.json",
		"second@testmail.json",
		"third@testmail.json",
		"fourth@testmail.json",
		"bcc@test.com",
	}
	cursor := 0
	lastCursor := len(expected)

	for recipient, err := range eachRecipient(map[string]any{
		"to": []any{
			// "first@testmail.com",
			"./recipients.yaml",
			"./recipients.cue",
			// "recipients.yaml",
			// "recipients.yaml",
			"bcc@test.com",
		},
	}, ".", os.DirFS("testdata")) {
		if err != nil {
			t.Fatal(err)
		}
		t.Log(cursor, recipient)
		email, ok := recipient[FieldNameEmail]
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
