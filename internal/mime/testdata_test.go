package mime

import (
	"net/http"
	"testing"

	"github.com/dkotik/mdsend/internal/media"
)

func TestEmbeddedContentType(t *testing.T) {
	if http.DetectContentType(media.Cat) != ContentTypeImageJPEG {
		t.Error("cat.jpg content type does not match")
	}
	if http.DetectContentType(media.Panda) != ContentTypeImageJPEG {
		t.Error("panda.jpg content type does not match")
	}
	if http.DetectContentType(media.Chamillion) != ContentTypeImageJPEG {
		t.Error("chamillion.jpg content type does not match")
	}
}
