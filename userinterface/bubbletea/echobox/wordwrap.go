package echobox

import (
	"strings"
	"unicode"
)

const noBreakSpace = 0xA0

func WordWrap(s string, lineLimit uint8) (result []string) {
	if lineLimit == 0 {
		return nil // otherwise word cut will not work and the last line will bleed over
	}

	var line, word, space strings.Builder
	var lineLength, wordLength, spaceLength uint8
	line.Grow(int(lineLimit) + 10)
	word.Grow(int(lineLimit) + 10)
	space.Grow(int(lineLimit))

	flushAllButLastLine := func() {
		if spaceLength > 0 { // deal with space
			if lineLength+spaceLength+wordLength > lineLimit {
				// space will be discarded
				result = append(result, line.String())
				line.Reset()
				lineLength = 0
			} else if lineLength > 0 { // keep space not at the front of a line
				line.WriteString(space.String())
				lineLength += spaceLength
			}
			space.Reset()
			spaceLength = 0
		}

		for _, char := range word.String() {
			line.WriteRune(char)
			lineLength++
			if lineLength == lineLimit {
				result = append(result, line.String())
				line.Reset()
				lineLength = 0
			}
		}
		word.Reset()
		wordLength = 0
	}

	for _, char := range s {
		if unicode.IsSpace(char) && char != noBreakSpace {
			if char == '\n' {
				flushAllButLastLine()
				if lineLength > 0 { // and the last line as well
					result = append(result, line.String())
					line.Reset()
					lineLength = 0
				}
				continue
			} else if wordLength > 0 {
				flushAllButLastLine()
			}
			space.WriteRune(char)
			spaceLength++
			continue
		}
		word.WriteRune(char)
		wordLength++
		if wordLength >= lineLimit { // force cut on super-long words
			flushAllButLastLine()
		}
	}

	flushAllButLastLine()
	if lineLength > 0 {
		result = append(result, line.String())
		line.Reset()
		lineLength = 0
	}
	return
}
