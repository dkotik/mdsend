package mime

import (
	"net/http"
	"testing"
)

func TestEmbeddedContentType(t *testing.T) {
	if http.DetectContentType(Cat) != ContentTypeImageJPEG {
		t.Error("cat.jpg content type does not match")
	}
	if http.DetectContentType(Panda) != ContentTypeImageJPEG {
		t.Error("panda.jpg content type does not match")
	}
	if http.DetectContentType(Chamillion) != ContentTypeImageJPEG {
		t.Error("chamillion.jpg content type does not match")
	}
}
