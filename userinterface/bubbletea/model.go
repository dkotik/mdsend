package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	padding  = 2
	maxWidth = 80
)

type msgDelivered struct {
	Address          string
	ConfirmationCode string
}

func (m model) readEvents() tea.Cmd {
	return func() tea.Msg {
		t := time.NewTimer(time.Millisecond * 1000)
		delivered := make([]msgDelivered, 0)

	loop:
		for {
			select {
			case <-t.C:
				break loop
			case msg := <-m.messages:
				delivered = append(delivered, msg)
				// case errors, can return those as well!
			}
		}
		return delivered
	}
}

type model struct {
	progress progress.Model
	messages chan (msgDelivered)
	report   []string
}

func (m model) Init() tea.Cmd {
	m.report = make([]string, 0)
	return m.readEvents()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil
	case progress.FrameMsg: // animate progress bar
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	case []msgDelivered:
		m.report = m.report[0:0]
		m.report = append(m.report,
			strings.Replace(fmt.Sprintf("%+v", msg), "} {", "\n  ", -1))

		if m.progress.Percent() == 1.0 {
			m.progress.Full = '▮'
		}

		cmd := m.progress.IncrPercent(0.01 * float64(len(msg)))
		return m, tea.Batch(
			cmd,
			m.readEvents(), // wait for next event
		)
	default:
		return m, nil
	}
}

func (m model) View() string {
	return strings.Join(m.report, "\n") + "\n\n" + strings.Repeat(" ", padding) + m.progress.View()
}
