package echobox

import (
	"mdsend/userinterface/bubbletea/scroll"
	"strings"
)

func (m *Model) messagesToLines(messages []EchoMsg) {
	if m.lineLength == 0 { // too early to render
		return
	}
	l := int(m.lineLength)

	for _, msg := range messages {
		for _, line := range WordWrap(msg.Message, m.lineLength) {
			m.lines = append(m.lines, msg.Style.Render(
				line+strings.Repeat("~", l-len(line))+scroll.DisplayTrack,
			))
		}
	}
}

func (m Model) View() string {
	window := m.cursor + m.height
	if window > len(m.lines) {
		window = len(m.lines)
	}

	// return spew.Sdump(m.lines[m.cursor:window])

	return strings.Join(m.lines[m.cursor:window], "\n")
}
