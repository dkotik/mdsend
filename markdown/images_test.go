package markdown

import (
	"bytes"
	"testing"
)

func TestImageRendering(t *testing.T) {
	md := New()

	var b bytes.Buffer
	md.Convert([]byte(`
# Test

![image](./local.jgp "kewl")
    `), &b)

	// t.Fatal(b.String())
}
