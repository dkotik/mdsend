package mime

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"mime"
	"net/textproto"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sebdah/goldie/v2"
)

// Per-Line Length: According to Internet Message Format, single lines in email headers must not exceed 998 characters, and are highly recommended to stay under 78 characters.

func TestNewHeaders(t *testing.T) {
	headers := []struct {
		Name  string
		Value string
	}{
		{Name: "Content-Type", Value: "text/plain"},
		{Name: "Content-Length", Value: "100"},
		{Name: "Content-Transfer-Encoding", Value: "base64"},
		{Name: "Content-Disposition", Value: "attachment; filename=test.txt"},
		{Name: "test", Value: "testValue"},
		{Name: "x", Value: "finalTestValue"},
		{Name: "a", Value: "firstTestValue"},
		{Name: "superLong", Value: strings.Repeat("--LONG--", 1000)},
		{Name: "unicode", Value: strings.Repeat("这是一句随机的话", 600)},
	}

	b := &bytes.Buffer{}
	var err error
	for _, h := range headers {
		if _, err = WriteHeader(b, h.Name, h.Value); err != nil {
			t.Errorf("WriteHeaders() = %v", err)
		}
	}
	goldie.New(t).Assert(t, "headers", b.Bytes())

	r := textproto.NewReader(bufio.NewReader(b))
	recovered, err := r.ReadMIMEHeader()
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("ReadMIMEHeader() = %v", err)
	}
	if len(recovered) != len(headers) {
		t.Errorf("headers count does not match: %d, want %d", len(recovered), len(headers))
	}
	for _, v := range headers {
		// if v.Value != recovered.Get(v.Name) {
		// 	// t.Fatalf("ReadMIMEHeader() = %v, want %v", recovered.Get(v.Name), v.Value)
		// 	t.Error("headers for the same key do not match:", v.Name)
		// }
		value := recovered.Get(v.Name)
		decoder := mime.WordDecoder{}
		value, err := decoder.DecodeHeader(value)
		if err != nil {
			t.Fatalf("DecodeHeader() = %v", err)
		}
		// if value != v.Value {
		// 	t.Errorf("DecodeHeader() = %v, want %v", value, v.Value)
		// }
		if len(value) > 72 {
			// wrapped header will contain extra spaces
			// that E-mail clients should drop to reconstruct
			// the full header value
			value = strings.ReplaceAll(value, " ", "")
		}
		if diff := cmp.Diff(v.Value, value); diff != "" {
			t.Fatalf("string mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestHeaderWithUTF8(t *testing.T) {
	var err error
	b := &bytes.Buffer{}
	if _, err = WriteHeader(b, "subject", "Замовлення"); err != nil {
		t.Fatalf("WriteHeader() = %v", err)
	}

	r := textproto.NewReader(bufio.NewReader(b))
	recovered, err := r.ReadMIMEHeader()
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("ReadMIMEHeader() = %v", err)
	}
	subject := recovered.Get("subject")
	if !strings.HasPrefix(subject, "=?UTF-8?b?") {
		t.Fatal("missing encoding prefix:", subject)
	}
	if !strings.HasSuffix(subject, "?=") {
		t.Fatal("missing encoding suffix:", subject)
	}

	decoder := &mime.WordDecoder{}
	subject, err = decoder.DecodeHeader(subject)
	if err != nil {
		t.Errorf("DecodeHeader() = %v", err)
	}
	if subject != "Замовлення" {
		t.Errorf("HeaderWithUTF8() = %v, want %v", subject, "Замовлення")
	}

	// b.Reset()
	// Header{
	// 	Name:  "from",
	// 	Value: "Офіс <info@example.com>",
	// }.WriteTo(b)

	// if b.String() != "From: =?utf-8?B?0J7RgdGC0LXQuQ==?= <info@example.com>" {
	// 	t.Errorf("HeaderWithUTF8() = %v, want %v", b.String(), "From: =?utf-8?B?0J7RgdGC0LXQuQ==?= <info@example.com>")
	// }
}
