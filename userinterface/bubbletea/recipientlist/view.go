package recipientlist

import (
	"fmt"
	"log"
	"mdsend/userinterface/bubbletea/scroll"
	"strings"
)

// var (
// 	appStyle = lipgloss.NewStyle().
// 			Foreground(lipgloss.Color("#43884a")).
// 			Background(lipgloss.Color("#3acf49"))
// 	helpStyle = lipgloss.NewStyle().
// 			Foreground(lipgloss.Color("#ff77a8")).
// 			Background(lipgloss.Color("#d64ce7"))
// )

func (m *Model) Render() (result []string) {
	result = make([]string, m.height)
	print := func(i int, s string) {
		if len(s) > m.width {
			result[i] = s[:m.width]
			return
		}
		result[i] = fmt.Sprintf("%*s", -m.width, s) // pad
		// result[i] = strings.Replace(fmt.Sprintf("%*s", -m.width, s), " ", "~", -1) // pad
	}

	for i := 0; i < m.height; i++ {
		offset := m.cursor + i
		index := offset / 2
		if index > len(m.Recipients)-1 {
			print(i, fmt.Sprintf("│end %d/%d            ", i+1, m.height))
			continue
		}

		if index < 0 {
			log.Fatal(index, offset, m.cursor)
		}

		rs := m.Recipients[index]
		if offset%2 == 0 {
			state := '?'
			switch rs.State {
			case DeliveryStateSent:
				state = '>'
			case DeliveryStateFailed:
				state = '!'
			}

			print(i, fmt.Sprintf("│%d/%d %c %s", i+1, m.height, state, rs.Name))
		} else {
			print(i, fmt.Sprintf("│%d/%d %s", i+1, m.height, rs.Address))
		}
	}

	if m.showControls {
		cutoff := m.width + 1
		scrolled := int(float32(m.cursor)/float32(m.cursorMaximum)*float32(m.height)) - 1
		for i := 0; i < m.height; i++ {
			if i == scrolled || i == 0 && scrolled < 1 {
				result[i] = result[i][:cutoff] + scroll.DisplayBar
			} else {
				result[i] = result[i][:cutoff] + scroll.DisplayTrack
			}
		}
	}
	return
}

func (m Model) View() string {
	return strings.Join(m.Render(), "\n")
}
