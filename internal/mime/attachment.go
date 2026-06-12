package mime

import (
	"context"
	"fmt"
	"io"
	"iter"

	"github.com/dkotik/mdsend"
)

var _ AttachmentRepository = (mdsend.Queue)(nil)

type AttachmentRepository interface {
	ListAttachments(ctx context.Context, letterID string) iter.Seq2[mdsend.Attachment, error]
}

type mockAttachmentRepository struct {
	attachments []mdsend.Attachment
}

func newMockAttachmentRepository(attachments ...mdsend.Attachment) AttachmentRepository {
	return mockAttachmentRepository{attachments: attachments}
}

func (m mockAttachmentRepository) ListAttachments(ctx context.Context, letterID string) iter.Seq2[mdsend.Attachment, error] {
	return func(yield func(mdsend.Attachment, error) bool) {
		for _, a := range m.attachments {
			if !yield(a, nil) {
				return
			}
		}
	}
}

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
