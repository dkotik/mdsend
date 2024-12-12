package loaders

import (
	"bytes"
	"testing"
)

type unsubTemplateFieldsForTesting struct {
	Name    string
	Address string
	ListID  string
}

func TestUnsubscribeLinkRendering(t *testing.T) {
	b := &bytes.Buffer{}
	tmpl, err := NewUnsubscribeLinkTemplate("https://remove.me/api/unsub?address={{urlQuery .Address }}&list={{ urlQuery .ListID }}")
	if err != nil {
		t.Fatal(err)
	}
	if err = tmpl.Execute(b, unsubTemplateFieldsForTesting{
		Address: "test@test.com",
		ListID:  "sdoifu3ur834rsdkflsdjf",
	}); err != nil {
		t.Fatal(err)
	}
	if b.String() != "https://remove.me/api/unsub?address=test%40test.com&list=sdoifu3ur834rsdkflsdjf" {
		t.Errorf("mismatched URL: expected %q but got %q", "https://remove.me/api/unsub?address=test%40test.com&list=sdoifu3ur834rsdkflsdjf", b.String())
	}
	b.Reset()

	tmpl, err = NewUnsubscribeLinkTemplate("https://remove.me/api/{{urlPath (base64 .Address) }}/{{ urlPath (base64 .ListID) }}")
	if err != nil {
		t.Fatal(err)
	}
	if err = tmpl.Execute(b, unsubTemplateFieldsForTesting{
		Address: "test@test.com",
		ListID:  "sdoifu3ur834rsdkflsdjf",
	}); err != nil {
		t.Fatal(err)
	}
	if b.String() != "https://remove.me/api/dGVzdEB0ZXN0LmNvbQ/c2RvaWZ1M3VyODM0cnNka2Zsc2RqZg" {
		t.Errorf("mismatched URL: expected %q but got %q", "https://remove.me/api/dGVzdEB0ZXN0LmNvbQ/c2RvaWZ1M3VyODM0cnNka2Zsc2RqZg", b.String())
	}
}
