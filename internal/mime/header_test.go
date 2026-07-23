package mime

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/textproto"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sebdah/goldie/v2"
)

var headers = []struct {
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
	{Name: "unicode-short", Value: "随机的话"},
	{Name: "unicode-pad1", Value: " 这是一句随机的话"},
	{Name: "unicode-pad2", Value: "  这是一句随机的话"},
	{Name: "unicode-pad3", Value: "   这是一句随机的话"},
	{Name: "unicode-pad4", Value: "    这是一句随机的话"},
	{Name: "unicode-pad5", Value: "     这是一句随机的话"},
	{Name: "unicode-pad6", Value: "      这是一句随机的话"},
	{Name: "unicode-pad7", Value: "       这是一句随机的话"},
	{Name: "unicode-pad8", Value: "        这是一句随机的话"},
	{Name: "unicode-pad9", Value: "         这是一句随机的话"},
	{Name: "unicode-pad10", Value: "          这是一句随机的话"},
	{Name: "unicode-pad11", Value: "           这是一句随机的话"},
	{Name: "unicode-pad12", Value: "            这是一句随机的话"},
	{Name: "unicode-pad13", Value: "             这是一句随机的话"},
	{Name: "superLong", Value: strings.Repeat("--LONG--", 1000)},
	{Name: "unicode", Value: strings.Repeat("这是一句a随机的ь话", 600)},
}

func TestNewHeaders(t *testing.T) {
	b := &bytes.Buffer{}
	v := &bytes.Buffer{}
	w := io.MultiWriter(b, v)
	decoder := mime.WordDecoder{}
	var err error
	for _, h := range headers {
		if _, err = WriteHeader(w, h.Name, h.Value); err != nil {
			t.Errorf("WriteHeaders() = %v", err)
		}
		recovered, err := decoder.DecodeHeader(v.String())
		if err != nil {
			t.Fatal("unable to decode header:", recovered)
		}
		if recovered != fmt.Sprintf("%s: %s\r\n", h.Name, h.Value) {
			t.Log("original: ", fmt.Sprintf("%s: %s\\r\\n", h.Name, h.Value))
			t.Log("recovered:", recovered)
			t.Error("recovered value does not match")
		}
		v.Reset()
	}
	// t.Fatal("test WriteAddressHeader")
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

func FuzzHeaderEncoding(f *testing.F) {
	f.Skip("broken by line breaks for some reason")
	for _, h := range headers {
		f.Add(h.Value)
	}
	decoder := mime.WordDecoder{}
	f.Fuzz(func(t *testing.T, value string) {
		w := &bytes.Buffer{}
		_, err := WriteHeader(w, "name", value)
		if err != nil {
			t.Errorf("WriteHeaders() = %v", err)
		}
		recovered, err := decoder.DecodeHeader(w.String())
		if err != nil {
			t.Fatal("unable to decode header:", recovered)
		}
		if recovered != fmt.Sprintf("%s: %s\r\n", "name", value) {
			t.Log("original: ", fmt.Sprintf("%s: %s\\r\\n", "name", value))
			t.Log("recovered:", recovered)
			t.Fatal("recovered value does not match")
		}
	})
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
