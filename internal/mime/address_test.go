package mime

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/mail"
	"net/textproto"
	"testing"

	"github.com/dkotik/mdsend/internal"
	"github.com/sebdah/goldie/v2"
)

func TestAddressHeader(t *testing.T) {
	b := &bytes.Buffer{}
	bcc := internal.MockAddresses
	err := WriteAddressHeader(b, "BCC", bcc...)
	if err != nil {
		t.Fatalf("WriteAddressHeader() = %v", err)
	}

	goldie.New(t).Assert(t, "headers_addr", b.Bytes())

	recovered, err := textproto.NewReader(bufio.NewReader(bytes.NewReader(b.Bytes()))).ReadMIMEHeader()
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("ReadMIMEHeader() = %v", err)
	}

	bccRaw := recovered.Get("BCC")
	if bccRaw == "" {
		t.Error("BCC header not found")
	}
	bccRecovered, err := mail.ParseAddressList(bccRaw)
	if err != nil {
		t.Fatalf("ParseAddressList() = %v", err)
	}
	if len(bccRecovered) != len(bcc) {
		t.Errorf("address count does not match: %d, want %d", len(bccRecovered), len(bcc))
	}
	for i, addr := range bccRecovered {
		if addr.Name != bcc[i].Name || addr.Address != bcc[i].Address {
			t.Errorf("address mismatch at index %d: %v, want %v", i, addr, bcc[i])
		}
	}

	scanner := bufio.NewScanner(b)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			t.Fatal("encountered an empty line")
		}
		if len(line) > LineLengthLimit {
			t.Log(line)
			t.Errorf("line too long: %d, want <= %d", len(line), LineLengthLimit)
		}
	}
}
