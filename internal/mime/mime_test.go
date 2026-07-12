package mime

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"os"
	"strings"
	"testing"

	"github.com/dkotik/mdsend/header"
)

func TestValidBoundaryGeneration(t *testing.T) {
	for range 100 {
		boundary := NewBoundary(entropy)
		// rfc2046#section-5.1.1
		if len(boundary) < 1 || len(boundary) > 70 {
			t.Fatal("mime: invalid boundary length")
		}
		end := len(boundary) - 1
		for i, b := range boundary {
			if 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' || '0' <= b && b <= '9' {
				continue
			}
			switch b {
			case '\'', '(', ')', '+', '_', ',', '-', '.', '/', ':', '=', '?':
				continue
			case ' ':
				if i != end {
					continue
				}
			}
			t.Fatal("mime: invalid boundary character")
		}
	}
}

type partDefinition struct {
	Headers []header.Header
	// Content  []byte
	Children []partDefinition
}

func (p partDefinition) Match(
	mh textproto.MIMEHeader,
	body io.Reader,
) func(*testing.T) {
	return func(t *testing.T) {
		wd := new(mime.WordDecoder)
		for _, h := range p.Headers {
			actual, err := wd.DecodeHeader(mh.Get(h.Name))
			if err != nil {
				t.Fatal("unable to decode header:", err)
			}
			if h.Value != actual {
				t.Log("A:", h.Value)
				t.Log("B:", actual)
				t.Fatal("header does not match")
			}
		}

		childrenCount := len(p.Children)
		if childrenCount == 0 {
			return
		}
		mediaType, params, err := mime.ParseMediaType(mh.Get(header.ContentType))
		if err != nil {
			t.Fatal(err)
		}
		boundary := params["boundary"]
		if boundary == "" {
			t.Fatal("boundary not found in multipart content type")
		}
		if !strings.HasPrefix(mediaType, "multipart/") {
			t.Fatal("expected multipart content type, got:", mediaType)
		}

		reader := multipart.NewReader(body, boundary)
		index := 0
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if index == childrenCount {
				t.Fatal("there are more children than expected")
			}
			t.Run(
				"validate child",
				p.Children[index].Match(part.Header, part),
			)
			index++
		}
		if index != childrenCount {
			t.Fatal("not all children were validated")
		}
	}
}

func ValidateMessageStructure(r io.Reader, m partDefinition) func(*testing.T) {
	return func(t *testing.T) {
		ml, err := mail.ReadMessage(r)
		if err != nil {
			t.Fatal("unable to parse electronic mail body:", err)
		}
		m.Match(textproto.MIMEHeader(ml.Header), ml.Body)(t)
	}
}

func TestFileEncoding(t *testing.T) {
	cat, err := os.ReadFile("testdata/image/cat.jpg")
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	e := NewEncoderBase64(b)
	if _, err = io.Copy(e, bytes.NewReader(cat)); err != nil {
		t.Fatal(err)
	}
	if err = e.Close(); err != nil {
		t.Fatal("unable to close base64 encoding")
	}
	if b.Len() == 0 {
		t.Fatal("nothing was written")
	}

	d := base64.NewDecoder(base64.StdEncoding, b)
	image := &bytes.Buffer{}
	if _, err = io.Copy(image, d); err != nil {
		t.Fatal(err)
	}
	if image.Len() == 0 {
		t.Fatal("nothing was written to the image buffer")
	}

	imageJPG, err := jpeg.Decode(image)
	if err != nil {
		t.Fatal(err)
	}
	bounds := imageJPG.Bounds()
	if bounds.Dx() != 618 {
		t.Error("width does not match:", bounds.Dx())
	}
	if bounds.Dy() != 750 {
		t.Error("height does not match:", bounds.Dy())
	}
}
