package internal

import (
	"net/http"
	"testing"

	"github.com/dkotik/mdsend/internal/mime"
)

func TestEmbeddedContentType(t *testing.T) {
	if http.DetectContentType(Cat) != mime.ContentTypeImageJPEG {
		t.Error("cat.jpg content type does not match")
	}
	if http.DetectContentType(Panda) != mime.ContentTypeImageJPEG {
		t.Error("panda.jpg content type does not match")
	}
	if http.DetectContentType(Chamillion) != mime.ContentTypeImageJPEG {
		t.Error("chamillion.jpg content type does not match")
	}
}
