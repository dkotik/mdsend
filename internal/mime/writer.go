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
	"github.com/dkotik/mdsend/header"
)

type Writer struct {
	attachments              AttachmentRepository
	cachedAttachments        map[string][]cachedAttachment
	cachedAttachmentContents map[string][]byte

	mixedBoundary   string
	textBoundary    string
	relatedBoundary string
}

func NewWriter(attachments AttachmentRepository, entropy *rand.Rand) Writer {
	if entropy == nil {
		now := uint64(time.Now().UnixNano())
		entropy = rand.New(rand.NewPCG(now/8, now))
	}
	if attachments == nil {
		attachments = newMockAttachmentRepository()
	}
	return Writer{
		attachments:              attachments,
		cachedAttachments:        make(map[string][]cachedAttachment),
		cachedAttachmentContents: make(map[string][]byte),
		mixedBoundary:            NewBoundary(entropy),
		textBoundary:             NewBoundary(entropy),
		relatedBoundary:          NewBoundary(entropy),
	}
}

func (w Writer) Write(
	ctx context.Context,
	out io.Writer,
	m mdsend.Message,
) (err error) {
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

	if err = WriteAddressHeader(out, header.From, m.From); err != nil {
		return err
	}
	if err = WriteAddressHeader(out, header.To, m.To); err != nil {
		return err
	}
	if _, err = WriteHeader(out, header.Subject, m.Subject); err != nil {
		return err
	}
	for _, header := range m.Headers {
		if _, err = WriteHeader(out, header.Name, header.Value); err != nil {
			return err
		}
	}
	if _, err = io.WriteString(out, header.MIMEVersion+": 1.0"+CRNL); err != nil {
		return err
	}

	if m.HTML == "" {
		if len(attachments) == 0 {
			return writeText(out, m.Text)
		}
		if err = w.WriteMixedBoundaryHeader(out); err != nil {
			return err
		}
		if _, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.mixedBoundary); err != nil {
			return err
		}
		if err = writeText(out, m.Text); err != nil {
			return err
		}
		for _, attachment := range attachments {
			_, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.mixedBoundary)
			if err != nil {
				return err
			}
			if err = attachment.WriteHeader(out); err != nil {
				return err
			}
			data, ok := w.cachedAttachmentContents[attachment.Hash]
			if !ok {
				return fmt.Errorf("attachment content not found: %s", attachment.Hash)
			}
			if _, err = io.Copy(out, bytes.NewReader(data)); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(out, "\r\n--%s--\r\n", w.mixedBoundary)
		return err
	}

	attachments, inline := SplitAttachments(m.HTML, attachments)
	if len(attachments) == 0 {
		return w.writeAlternativeWithAttachments(out, m.Text, m.HTML, w.textBoundary, inline)
	}

	if err = w.WriteMixedBoundaryHeader(out); err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, CRNL+"--%s\r\n", w.mixedBoundary)
	if err != nil {
		return err
	}

	if len(inline) == 0 {
		if err = writeAlternative(out, m.Text, m.HTML, w.textBoundary); err != nil {
			return err
		}
	} else {
		if _, err = out.Write([]byte(header.ContentType + `: multipart/alternative;` + CRNL + ` boundary="`)); err != nil {
			return err
		}
		if _, err = out.Write([]byte(w.textBoundary)); err != nil {
			return err
		}
		if _, err = out.Write([]byte(`"` + CRNL + CRNL)); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "--%s\r\n", w.textBoundary)
		if err != nil {
			return err
		}
		if err = writeText(out, m.Text); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.textBoundary)
		if err != nil {
			return err
		}
		if err = w.WriteRelatedBoundaryHeader(out); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.relatedBoundary)
		if err != nil {
			return err
		}
		if err = writeHTML(out, m.HTML); err != nil {
			return err
		}
		for _, attachment := range inline {
			_, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.relatedBoundary)
			if err != nil {
				return err
			}
			if err = attachment.WriteInlineHeader(out, attachment.CanonicalContentID); err != nil {
				return err
			}
			data, ok := w.cachedAttachmentContents[attachment.Hash]
			if !ok {
				return fmt.Errorf("attachment content not found: %s", attachment.Hash)
			}
			if _, err = io.Copy(out, bytes.NewReader(data)); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(out, "\r\n--%s--\r\n", w.relatedBoundary)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "\r\n--%s--\r\n", w.textBoundary)
		if err != nil {
			return err
		}
	}

	for _, attachment := range attachments {
		_, err = fmt.Fprintf(out, "\r\n--%s\r\n", w.mixedBoundary)
		if err != nil {
			return err
		}
		if err = attachment.WriteHeader(out); err != nil {
			return err
		}
		data, ok := w.cachedAttachmentContents[attachment.Hash]
		if !ok {
			return fmt.Errorf("attachment content not found: %s", attachment.Hash)
		}
		if _, err = io.Copy(out, bytes.NewReader(data)); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(out, "\r\n--%s--\r\n", w.mixedBoundary)
	return err
}
