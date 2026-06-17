package mime

import (
	"fmt"
	"regexp"
	"strings"
)

const InlineReferenceLimit = 64

var reInlineSourceCapture = regexp.MustCompile(`src="cid:([^"]+)"`)

type inlineAttachment struct {
	cachedAttachment
	CanonicalContentID string
}

func SplitAttachments(html string, attachments []cachedAttachment) (normal []cachedAttachment, inline []inlineAttachment) {
	hashToDomain := make(map[string]string)
	for _, m := range reInlineSourceCapture.FindAllStringSubmatch(html, InlineReferenceLimit) {
		hash, domain, ok := strings.Cut(m[1], "@")
		if !ok || hash == "" || domain == "" {
			continue
		}
		hashToDomain[hash] = domain
	}
	normal = make([]cachedAttachment, 0, len(attachments))
	inline = make([]inlineAttachment, 0, 2)

	for _, a := range attachments {
		domain, ok := hashToDomain[a.Hash]
		if !ok {
			normal = append(normal, a)
			continue
		}
		inline = append(inline, inlineAttachment{
			cachedAttachment:   a,
			CanonicalContentID: fmt.Sprintf("<%s@%s>", a.Hash, domain),
		})
	}
	return normal, inline
}
