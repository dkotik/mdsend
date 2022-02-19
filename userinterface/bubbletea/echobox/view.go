package echobox

import "strings"

func (m *Model) messagesToLines(messages []EchoMsg) {
	for _, msg := range messages {
		for _, line := range WordWrap(msg.Message, m.lineLength) {
			m.lines = append(m.lines, msg.Style.Render(line))
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
