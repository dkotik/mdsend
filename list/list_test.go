package list

import (
	"context"
	"testing"
)

func TestCreateAddSave(t *testing.T) {
	ctx := context.Background()
	l, err := NewList(
		ctx,
		"testdata/source.json",
		"testdata/source.yaml",
	)
	if err != nil {
		t.Error(err)
	}

	if len(l.contacts) != 8 {
		t.Errorf("expected %d contacts, but got %d", 8, len(l.contacts))
	}

	contact, err := l.AddContact(l.sources[0])
	if err != nil {
		t.Error(err)
	}
	if err = contact.SetName(ctx, "some test contact"); err != nil {
		t.Error(err)
	}

	// if err = l.ReplaceSource(l.sources[0], sourceJSON("testdata/source3.json")); err != nil {
	// 	t.Error(err)
	// }
	// if l.sources[0].Location() != contact.source.Location() {
	// 	t.Error("source replacement failed", l.sources[0].Location(), contact.source.Location())
	// }

	if err = contact.Delete(ctx); err != nil {
		t.Error(err)
	}

	// if err = l.ReplaceSource(l.sources[1], sourceYAML("testdata/source2.yaml")); err != nil {
	// 	t.Error(err)
	// }

	if err = l.contacts[6].SetName(ctx, "ooga booga"); err != nil {
		t.Error(err)
	}
	if err = l.contacts[6].SetEmail(ctx, "ooga@booga.test"); err != nil {
		t.Error(err)
	}
}
