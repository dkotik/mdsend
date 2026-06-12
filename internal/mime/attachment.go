package mime

import (
	"fmt"
	"io"
)

type cachedAttachment struct {
	Name        string
	Hash        string
	ContentType string
}

func (a cachedAttachment) WriteHeader(w io.Writer) (err error) {
	if _, err = WriteHeader(w, HeaderContentType, a.ContentType); err != nil {
		return err
	}
	if _, err = WriteHeader(w, HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, escapeQuotes(a.Name))); err != nil {
		return err
	}
	if _, err = WriteHeader(w, HeaderContentTransferEncoding, `base64`); err != nil {
		return err
	}
	// _, err = io.WriteString(w, CRNL)
	return nil
}

func (a cachedAttachment) WriteInlineHeader(w io.Writer, contentID string) (err error) {
	if _, err = WriteHeader(w, HeaderContentType, a.ContentType); err != nil {
		return err
	}
	if _, err = WriteHeader(w, HeaderContentDisposition, fmt.Sprintf(`inline; filename="%s"`, escapeQuotes(a.Name))); err != nil {
		return err
	}
	if _, err = WriteHeader(w, HeaderContentTransferEncoding, `base64`); err != nil {
		return err
	}
	if _, err = WriteHeader(w, HeaderContentID, contentID); err != nil {
		return err
	}
	// _, err = io.WriteString(w, CRNL)
	return nil
}
