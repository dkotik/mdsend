package loader

import "testing"

func TestRecipientListFromYAML(t *testing.T) {
	expected := []string{
		"first@testmail.com",
		"second@testmail.com",
		"third@testmail.com",
		"fourth@testmail.com",
	}
	cursor := 0
	for recipient, err := range eachRecipientFromFileYAML("../internal/testdata/recipients.yaml") {
		t.Log(recipient)
		if err != nil {
			t.Fatal(err)
		}
		email, ok := recipient[EmailKey]
		if !ok {
			t.Fatal("recipient does not contain an email address")
		}
		if expected[cursor] != email {
			t.Fatal("unexpected email:", email, "vs", expected[cursor])
		}
		cursor++
	}
}

func TestRecipientList(t *testing.T) {
	expected := []string{
		"first@testmail.com",
		"second@testmail.com",
		"third@testmail.com",
		"fourth@testmail.com",
		"bcc@test.com",
	}
	cursor := 0
	lastCursor := len(expected)

	m, err := NewMessage("../internal/testdata/pass/a.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Message ID:", m.ID)
	for recipient, err := range m.EachRecipient() {
		t.Log(recipient)
		if err != nil {
			t.Fatal(err)
		}
		email, ok := recipient[EmailKey]
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
		t.Fatal("too few contacts loaded")
	}
}
