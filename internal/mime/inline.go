package mime

import (
	"fmt"
	"regexp"
	"strings"
)

const InlineReferenceLimit = 64

var reInlineSourceCapture = regexp.MustCompile(`src="cid:([^"]+)"`)

type contentID struct {
	Hash      string
	Canonical string
}

type InlineReferences struct {
	ids []contentID
}

func FindInlineReferences(html string) InlineReferences {
	matches := reInlineSourceCapture.FindAllStringSubmatch(html, InlineReferenceLimit)
	ids := make([]contentID, 0, len(matches))
	for _, m := range matches {
		hash, domain, ok := strings.Cut(m[1], "@")
		if !ok || hash == "" || domain == "" {
			continue
		}
		ids = append(ids, contentID{
			Hash:      hash,
			Canonical: fmt.Sprintf("<%s@%s>", hash, domain),
		})
	}
	return InlineReferences{ids: ids}
}

func (r InlineReferences) MatchContentID(hash string) string {
	for _, id := range r.ids {
		if id.Hash == hash {
			return id.Canonical
		}
	}
	return ""
}

func (r InlineReferences) Count() int {
	return len(r.ids)
}
