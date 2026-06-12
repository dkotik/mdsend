package mime

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"io"
	"os"
	"testing"
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

func TestMessageStructure(t *testing.T) {
	// mime.TypeByExtension(ext string)
	// mime.ParseMediaType(v string)
	// r := multipart.NewReader(r io.Reader, boundary string)
}

func TestFileEncoding(t *testing.T) {
	cat, err := os.ReadFile("../testdata/cat.jpg")
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
