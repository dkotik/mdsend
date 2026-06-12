package mime

import (
	"context"
	"iter"

	"github.com/dkotik/mdsend"
)

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
