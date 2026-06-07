package mime

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/mail"
	"net/textproto"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func TestAddressHeader(t *testing.T) {
	b := &bytes.Buffer{}
	bcc := []mail.Address{
		mail.Address{Name: "Joe Crazy", Address: "crazyjoe@test.com"},
		mail.Address{Name: "Перший", Address: "first@test.com"},
		mail.Address{Name: "Вторій", Address: "second@test.com"},
		mail.Address{Name: "Третій", Address: "third@test.com"},
		mail.Address{Name: "Четвертий", Address: "fourth@test.com"},
		mail.Address{Name: "П'ятий", Address: "fifth@test.com"},
		mail.Address{Name: "Шостий", Address: "sixth@test.com"},
		mail.Address{Name: "Сьомий", Address: "seventh@test.com"},
		mail.Address{Name: "Восьмий", Address: "eighth@test.com"},
		mail.Address{Name: "Дев'ятий", Address: "ninth@test.com"},
		mail.Address{Name: "Joe", Address: "crazyjoe@test.com"},
		mail.Address{Name: "", Address: "empty@test.com"},
		mail.Address{Name: "Вторій", Address: "second@test.com"},
		mail.Address{Name: "Третій", Address: "third@test.com"},
		mail.Address{Name: "Четвертий", Address: "fourth@test.com"},
		mail.Address{Name: "П'ятий", Address: "fifth@test.com"},
		mail.Address{Name: "Шостий", Address: "sixth@test.com"},
		mail.Address{Name: "Сьомий", Address: "seventh@test.com"},
		mail.Address{Name: "Восьмий", Address: "eighth@test.com"},
		mail.Address{Name: "Дев'ятий", Address: "ninth@test.com"},
	}
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
