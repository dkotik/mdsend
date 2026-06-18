package markdown

// var reHorizontalRule = regexp.MustCompile(`(?m)^[ \t]*(?:([-_*])\s*)\1\s*\1(?:\s*\1)*$`)

const falseChar = byte('@')

func SplitOnLastHorizontalRule(data []byte) (most, tail []byte, ok bool) {
	first := falseChar
	count := 0
	lastLineIndex := 0
	cutBegin := 0
	cutEnd := 0
	for i, c := range data {
		switch c {
		case ' ', '\t':
		// ignore whitespace
		case '-', '_', '*':
			if count == 0 {
				first = c
			} else if c != first {
				first = falseChar
			}
			count++
		case '\n', '\r':
			if first != falseChar && count >= 3 {
				cutBegin = lastLineIndex + 1
				cutEnd = i + 1
			}
			count = 0
			first = falseChar
			lastLineIndex = i
		default:
			first = falseChar
		}
	}

	if cutBegin == 0 {
		return data, nil, false
	}
	for _, c := range data[cutEnd:] {
		// drain whitespace from the beginnig of the tail
		switch c {
		case ' ', '\t', '\n', '\r':
			cutEnd++
		default:
			goto done
		}
	}

done:
	return data[:cutBegin], data[cutEnd:], true
}
