package html

import (
	"bytes"
	"iter"
)

func getPrefixIndexIgnoreWhitespace(b, prefix []byte) int {
	cursor := 0
	m := prefix[cursor]
	for i, c := range b {
		switch c {
		case ' ', '\t', '\n', '\r':
			if cursor > 0 {
				return -1
			}
			continue
		case m:
			cursor++
			if cursor == len(prefix) {
				return i - cursor + 1
			}
			m = prefix[cursor]
		default:
			return -1
		}
	}
	return -1
}

func getPrefixIndexIgnoreWhitespaceForContentID(b []byte) int {
	return getPrefixIndexIgnoreWhitespace(b, []byte(`cid:`))
}

func EachSourceWithContentID(b []byte) iter.Seq[string] {
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
				if i < 0 {
					continue // no matching quote
				}
				if prefixIndex := getPrefixIndexIgnoreWhitespaceForContentID(b[:i]); prefixIndex >= 0 {
					if !yield(string(b[prefixIndex+4 : i])) {
						return
					}
				}
				b = b[i+1:] // consume `'`
				continue
			case '"':
				b = b[i+1:] // jump over `src='`
				i = bytes.IndexByte(b, '"')
				if i < 0 {
					continue // no matching quote or `cid:` prefix
				}
				if prefixIndex := getPrefixIndexIgnoreWhitespaceForContentID(b[:i]); prefixIndex >= 0 {
					if !yield(string(b[prefixIndex+4 : i])) {
						return
					}
				}
				b = b[i+1:] // consume `"`
				continue
			default:
				b = b[i:]
				prefixIndex := getPrefixIndexIgnoreWhitespaceForContentID(b)
				if prefixIndex == -1 {
					b = b[1:]
					continue // no matching `cid:` prefix
				}
				b = b[prefixIndex+4:]
			}

		findBreak:
			for i, c = range b {
				switch c {
				case '\'', '"', ' ', '\\', '/', '>', '\t', '\n', '<':
					if !yield(string(b[:i])) {
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
