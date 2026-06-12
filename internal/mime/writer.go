package mime

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand/v2"
	"time"

	"github.com/dkotik/mdsend"
)

type Writer struct {
	w                        io.Writer
	attachments              AttachmentRepository
	cachedAttachments        map[string][]cachedAttachment
	cachedAttachmentContents map[string][]byte

	mixedBoundary string
	textBoundary  string
}

func NewWriter(w io.Writer, attachments AttachmentRepository, entropy *rand.Rand) Writer {
	if entropy == nil {
		now := uint64(time.Now().UnixNano())
		entropy = rand.New(rand.NewPCG(now/8, now))
	}
	if attachments == nil {
		attachments = newMockAttachmentRepository()
	}
	return Writer{
		w:                        w,
		attachments:              attachments,
		cachedAttachments:        make(map[string][]cachedAttachment),
		cachedAttachmentContents: make(map[string][]byte),
		mixedBoundary:            NewBoundary(entropy),
		textBoundary:             NewBoundary(entropy),
	}
}

func (w Writer) Write(ctx context.Context, m mdsend.Dispatch) (err error) {
	attachments, ok := w.cachedAttachments[m.LetterID]
	if !ok {
		for attachment, err := range w.attachments.ListAttachments(ctx, m.LetterID) {
			if err != nil {
				return fmt.Errorf("attachment retrieval error: %w", err)
			}
			attachments = append(attachments, cachedAttachment{
				Name:        attachment.Name,
				Hash:        attachment.Hash,
				ContentType: attachment.ContentType,
			})
			b := bytes.NewBuffer(make([]byte, 0, base64.StdEncoding.EncodedLen(len(attachment.Content))+len(CRNL)))
			_, _ = io.WriteString(b, CRNL)
			encoder := base64.NewEncoder(base64.StdEncoding, &lineWrapper{w: b})
			if _, err = io.Copy(encoder, bytes.NewReader(attachment.Content)); err != nil {
				return err
			}
			w.cachedAttachmentContents[attachment.Hash] = b.Bytes()
		}
		w.cachedAttachments[m.LetterID] = attachments
	}

	if err = WriteAddressHeader(w.w, HeaderFrom, m.From); err != nil {
		return err
	}
	if err = WriteAddressHeader(w.w, HeaderTo, m.To); err != nil {
		return err
	}
	if _, err = WriteHeader(w.w, HeaderSubject, m.Subject); err != nil {
		return err
	}
	for _, header := range m.Headers {
		if _, err = WriteHeader(w.w, header.Name, header.Value); err != nil {
			return err
		}
	}
	if _, err = io.WriteString(w.w, HeaderMIMEVersion+": 1.0"+CRNL); err != nil {
		return err
	}

	if m.HTML == "" {
		if len(attachments) == 0 {
			return writeText(w.w, m.Text)
		}
		if err = w.WriteMixedBoundaryHeader(w.w); err != nil {
			return err
		}
		if _, err = fmt.Fprintf(w.w, "--%s\r\n", w.mixedBoundary); err != nil {
			return err
		}
		if err = writeText(w.w, m.Text); err != nil {
			return err
		}
		for _, attachment := range attachments {
			_, err = fmt.Fprintf(w.w, "\r\n--%s\r\n", w.mixedBoundary)
			if err != nil {
				return err
			}
			if err = attachment.WriteHeader(w.w); err != nil {
				return err
			}
			data, ok := w.cachedAttachmentContents[attachment.Hash]
			if !ok {
				return fmt.Errorf("attachment content not found: %s", attachment.Hash)
			}
			if _, err = io.Copy(w.w, bytes.NewReader(data)); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(w.w, "\r\n--%s--\r\n", w.mixedBoundary)
		return err
	}

	if len(attachments) == 0 {
		return writeAlternative(w.w, m.Text, m.HTML, w.textBoundary)
	}

	inlineReferences := FindInlineReferences(m.HTML)

	// if _, err = io.WriteString(w.w, CRNL); err != nil {
	// 	return err
	// }

	if inlineReferences.Count() == 0 {
		if err = w.WriteMixedBoundaryHeader(w.w); err != nil {
			return err
		}
	} else {
		if err = w.WriteRelatedBoundaryHeader(w.w); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(w.w, CRNL+"--%s\r\n", w.mixedBoundary)
	if err != nil {
		return err
	}
	if err = writeAlternative(w.w, m.Text, m.HTML, w.textBoundary); err != nil {
		return err
	}
	for _, attachment := range attachments {
		_, err = fmt.Fprintf(w.w, "\r\n--%s\r\n", w.mixedBoundary)
		if err != nil {
			return err
		}
		cid := inlineReferences.MatchContentID(attachment.Hash)
		if cid == "" {
			if err = attachment.WriteHeader(w.w); err != nil {
				return err
			}
		} else {
			if err = attachment.WriteInlineHeader(w.w, cid); err != nil {
				return err
			}
		}
		data, ok := w.cachedAttachmentContents[attachment.Hash]
		if !ok {
			return fmt.Errorf("attachment content not found: %s", attachment.Hash)
		}
		if _, err = io.Copy(w.w, bytes.NewReader(data)); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(w.w, "\r\n--%s--\r\n", w.mixedBoundary)
	return err
}
