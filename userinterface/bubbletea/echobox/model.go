package echobox

import (
	"mdsend/userinterface/bubbletea/scroll"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	EchoMsg struct {
		Message string
		Style   lipgloss.Style
	}
	ClearMsg string
)

type Model struct {
	cursor     int
	width      int
	height     int
	messages   []EchoMsg
	lines      []string
	lineLength uint8
	Momentum   *scroll.Momentum
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.lineLength = uint8(msg.Width - 1)
		if len(m.messages) > 0 {
			m.lines = nil
			m.messagesToLines(m.messages)
			if l := len(m.lines) - m.height; m.cursor > l {
				if l < 0 {
					m.cursor = 0
				} else {
					m.cursor = l // shrunk buffer
				}
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			speed := m.Momentum.Up()
			m.cursor += -speed * speed
			if m.cursor < 0 {
				m.cursor = 0
			}
			// cmd := m.triggerControlsActivation()
			return m, nil
		case "down", "j":
			lines := len(m.lines)
			if lines > m.height {
				speed := m.Momentum.Down()
				m.cursor += speed * speed
				if max := lines - m.height; m.cursor > max {
					m.cursor = max
				}
			}

			// m.cursor += m.getMomentum()
			// if m.cursor > m.cursorMaximum {
			// 	m.cursor = m.cursorMaximum
			// }
			// cmd := m.triggerControlsActivation()
			return m, nil
		}
	case EchoMsg:
		return m, func() tea.Msg {
			return []EchoMsg{msg}
		}
	case []EchoMsg:
		m.messages = append(m.messages, msg...)
		m.messagesToLines(msg)
		return m, nil
	case ClearMsg:
		m.lines = nil
		m.messages = nil
		return m, nil
	case scroll.Message:
		switch msg {
		case scroll.ToBottom:
			if l := len(m.lines); l > m.height {
				m.cursor = len(m.lines) - m.height
			}
		}
	}
	return m, nil
}
