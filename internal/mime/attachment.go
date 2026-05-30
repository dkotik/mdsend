package mime

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/textproto"
)

type Attachment struct {
	Name        string
	ContentType string
	Content     []byte
}

func (a Attachment) Encode(mime *MultipartWriter) error {
	header := textproto.MIMEHeader{}
	header.Set(`Content-Type`, a.ContentType)
	// header.Set(`Content-Type`, http.DetectContentType(fileHead[:fileHeadBytesRead]))
	// if inline {
	// 	header.Set(`Content-Disposition`, fmt.Sprintf(`inline; filename=%q`, fname))
	// 	header.Set(`Content-ID`, fmt.Sprintf(`<%s>`, fname))
	// }
	header.Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename=%q`, a.Name))
	header.Set(`Content-Transfer-Encoding`, `base64`)
	partWriter, err := mime.CreatePart(header)
	if err != nil {
		return err
	}
	w := base64.NewEncoder(base64.StdEncoding, &lineWrapper{w: partWriter})
	defer w.Close()
	_, err = io.Copy(w, bytes.NewReader(a.Content))
	return err
}

type EmbeddedAttachment Attachment
