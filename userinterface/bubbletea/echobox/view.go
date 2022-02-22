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
				line+strings.Repeat("~", l-len(line))+scroll.DisplayTrack,
			))
		}
	}
}

func (m *Model) Render() []string {
	window := m.cursor + m.height
	if window > len(m.lines) {
		window = len(m.lines)
	}
	return m.lines[m.cursor:window]
}

func (m Model) View() string {
	return strings.Join(m.Render(), "\n")
}
