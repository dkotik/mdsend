package mime

import (
	"net/http"
	"testing"

	"github.com/dkotik/mdsend/internal"
)

func TestEmbeddedContentType(t *testing.T) {
	if http.DetectContentType(internal.Cat) != ContentTypeImageJPEG {
		t.Error("cat.jpg content type does not match")
	}
	if http.DetectContentType(internal.Panda) != ContentTypeImageJPEG {
		t.Error("panda.jpg content type does not match")
	}
	if http.DetectContentType(internal.Chamillion) != ContentTypeImageJPEG {
		t.Error("chamillion.jpg content type does not match")
	}
}
