package echobox

import (
	"fmt"
	"strings"
	"unicode"
)

const noBreakSpace = 0xA0

func (m *Model) wordWrap(s string) (result []string) {
	var line, word strings.Builder
	var cursor uint8
	line.Grow(int(m.lineLength) + 10)
	word.Grow(int(m.lineLength) + 10)

	flush := func() {
		line.WriteString(word.String())
		result = append(result,
			fmt.Sprintf("%*s", -m.lineLength, line.String()), // pad with spaces
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
					if uint8(line.Len())+cursor < m.lineLength {
						line.WriteString(word.String()) // only flush the word
						word.Reset()
						cursor = 0
					} else {
						flush() // flush the whole line
					}
					continue
				}
			}
			word.WriteRune(char)
			cursor++
			if cursor == m.lineLength {
				flush()
			}
		}
	}

	if cursor > 0 {
		flush() // write left-overs
	}

	return
}
