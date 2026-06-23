package markdown

import (
	"bytes"
	"testing"
)

func TestSplitOnLastHorizontalRule(t *testing.T) {
	data := []byte(`---
title: Test
---

* * *

_ _ _ _


tail`)
	most, tail, found := SplitOnLastHorizontalRule(data)
	if !found {
		t.Fatal("expected to find a horizontal rule")
	}
	if len(most) == 0 {
		t.Fatal("front of content was not found")
	}
	if len(tail) == 0 {
		t.Fatal("tail of content was not found")
	}
	if !bytes.Equal(tail, []byte(`tail`)) {
		t.Fatal("tail of content does not match expected")
	}
}
