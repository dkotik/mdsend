package recipientlist

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	DeliveryState uint8
	// TickMsg         time.Time
	HideControlsMsg time.Time
)

const (
	DeliveryStatePending DeliveryState = iota
	DeliveryStateSent
	DeliveryStateFailed

	delay = time.Millisecond * 1000
)

type Model struct {
	cursor                int
	cursorMaximum         int
	scrollMomentum        int
	scrollMomentumMaximum int
	width                 int
	height                int
	Recipients            []Recipient
	ControlsTimer         chan (*struct{})
	showControls          bool
}

func (m *Model) triggerControlsActivation() tea.Cmd {
	m.showControls = true
	return func() tea.Msg {
		select { // attempt to send a message to the channel
		case m.ControlsTimer <- &struct{}{}:
			return nil
		default:
			return nil
		}
	}
}

func (m *Model) listenForControlsActivation() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.ControlsTimer
		if msg == nil { // channel closed
			return nil
		}

		timer := time.NewTimer(delay)
		for {
			select {
			case msg := <-m.ControlsTimer:
				if msg == nil { // channel closed
					return nil
				}
				if !timer.Stop() {
					<-timer.C // drain channel, see docs
				}
				timer.Reset(delay)
			case t := <-timer.C:
				return HideControlsMsg(t)
			}
		}
	}
}

func (m Model) Init() tea.Cmd {
	// m.controlsTimer = make(chan (*struct{}), 1) // TODO: replace with external channel that closes? or add a closer interface
	// m.scrollMomentumMaximum = len(m.recipients)/5 + 30
	cmd := m.listenForControlsActivation()
	return cmd
}

func (m *Model) getMomentum() int {
	if !m.showControls {
		m.scrollMomentum = 1 // reset
	} else if m.scrollMomentum < m.scrollMomentumMaximum {
		m.scrollMomentum += m.scrollMomentumMaximum / 3
	}
	return m.scrollMomentum
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.cursorMaximum = len(m.Recipients)*2 - m.height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			m.cursor -= m.getMomentum()
			if m.cursor < 0 {
				m.cursor = 0
			}
			cmd := m.triggerControlsActivation()
			return m, cmd
		case "down", "j":
			m.cursor += m.getMomentum()
			if m.cursor > m.cursorMaximum {
				m.cursor = m.cursorMaximum
			}
			cmd := m.triggerControlsActivation()
			return m, cmd
		}
	case HideControlsMsg:
		m.showControls = false
		cmd := m.listenForControlsActivation()
		return m, cmd
	}
	return m, nil
}
