package mime

import (
	"bytes"
	"fmt"
	"iter"
	"strings"
)

// const InlineReferenceLimit = 64
// var reInlineSourceCapture = regexp.MustCompile(`src="cid:([^"]+)"`)

func eachInlineReferenceInHTML(b []byte) iter.Seq[string] {
	return func(yield func(string) bool) {
		var (
			i int
			c byte
		)

		for {
			i = bytes.Index(b, []byte(`src=`))
			if i < 0 {
				return // no more source references
			}
			if len(b) < 10 { // len(`src=cid:@`)
				return // end of string
			}
			i = i + 4 // len(`src=?`)
			switch b[i] {
			case '\'':
				// panic("no more source references")
				b = b[i+1:] // jump over `src='`
				i = bytes.IndexByte(b, '\'')
				if i < 0 || !bytes.HasPrefix(b, []byte(`cid:`)) {
					continue // no matching quote or `cid:` prefix
				}
				if !yield(string(b[4:i])) {
					return
				}
				b = b[i+1:] // consume `'`
				continue
			case '"':
				b = b[i+1:] // jump over `src='`
				i = bytes.IndexByte(b, '"')
				if i < 0 || !bytes.HasPrefix(b, []byte(`cid:`)) {
					continue // no matching quote or `cid:` prefix
				}
				if !yield(string(b[4:i])) {
					return
				}
				b = b[i+1:] // consume `"`
				continue
			case 'c':
				b = b[i:]
				if !bytes.HasPrefix(b, []byte(`cid:`)) {
					// panic(fmt.Sprintf("%d, %s", i, b))
					// b = b[i:] // len(`src=c`)
					continue // does not start with cid
				}
				// b = b[i+4:] // len(`cid:`)
			default:
				b = b[i+1:] // len(`src=?`)
				continue    // no possibility of cid prefix
			}

		findBreak:
			for i, c = range b {
				switch c {
				case '\'', '"', ' ', '\\', '/', '>', '\t', '\n', '<':
					if !yield(string(b[4:i])) {
						return
					}
					// i++
					b = b[i+1:]
					break findBreak
				}
			}
		}
	}
}

type inlineAttachment struct {
	cachedAttachment
	CanonicalContentID string
}

func SplitAttachments(html string, attachments []cachedAttachment) (normal []cachedAttachment, inline []inlineAttachment) {
	hashToDomain := make(map[string]string)
	for m := range eachInlineReferenceInHTML([]byte(html)) {
		hash, domain, ok := strings.Cut(m, "@")
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
