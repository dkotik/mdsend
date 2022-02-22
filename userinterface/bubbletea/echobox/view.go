package echobox

import (
	"mdsend/userinterface/bubbletea/scroll"
	"strings"
)

func (m *Model) messagesToLines(messages []EchoMsg) {
	l := int(m.lineLength)

	for _, msg := range messages {
		for _, line := range WordWrap(msg.Message, m.lineLength) {
			m.lines = append(m.lines, msg.Style.Render(
				line+strings.Repeat("~", l-len(line)),
			))
		}
	}
}

func (m *Model) Render() []string {
	window := m.cursor + m.height
	if window > len(m.lines) {
		window = len(m.lines)
	}

	scrolled := int(float32(m.cursor)/float32(len(m.lines)-m.height)*float32(m.height)) - 1
	result := make([]string, 0, m.height)
	for i, line := range m.lines[m.cursor:window] {
		if i == scrolled || i == 0 && scrolled < 1 {
			result = append(result, line+scroll.DisplayBar)
		} else {
			result = append(result, line+scroll.DisplayTrack)
		}

	}
	// spew.Dump(result)
	// panic("done")
	return result
	// return m.lines[m.cursor:window]
}

func (m Model) View() string {
	return strings.Join(m.Render(), "\n")
}
