package echobox

import (
	"strings"
	"unicode"
)

const noBreakSpace = 0xA0

func WordWrap(s string, lineLength uint8) (result []string) {
	var line, word strings.Builder
	var cursor uint8
	line.Grow(int(lineLength) + 10)
	word.Grow(int(lineLength) + 10)

	flush := func() {
		line.WriteString(word.String())
		result = append(result,
			// fmt.Sprintf("%*s", lineLength, line.String()), // pad with spaces
			line.String(),
		)
		word.Reset()
		line.Reset()
		cursor = 0
	}

	for _, char := range s {
		if char == '\n' {
			flush()
		} else {
			if unicode.IsSpace(char) && char != noBreakSpace {
				if cursor > 0 {
					if uint8(line.Len())+cursor <= lineLength {
						line.WriteString(word.String()) // only flush the word
						word.Reset()
						word.WriteRune(char)
						cursor = 1
					} else {
						result = append(result,
							// fmt.Sprintf("%*s", lineLength, line.String()), // pad with spaces
							line.String(),
						)
						line.Reset()
						line.WriteString(strings.TrimSpace(word.String()))
						line.WriteRune(char)
						word.Reset()
						cursor = 0
					}
					continue
				}
			}
			word.WriteRune(char)
			cursor++
			if cursor == lineLength {
				flush()
			}
		}
	}

	if cursor > 0 {
		flush() // write left-overs
	}

	return
}
