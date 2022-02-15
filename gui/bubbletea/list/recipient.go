package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle()
)

type Recipient struct {
	Name    string
	Address string
	Fields  map[string]interface{}
}

type RecipientList struct {
	cursor     int
	width      int
	height     int
	recipients []Recipient
}

func (r RecipientList) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (r RecipientList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width, r.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return r, tea.Quit
		case "up", "k":
			if r.cursor > 0 {
				r.cursor--
			}
		case "down", "j":
			if r.cursor < len(r.recipients)-1 {
				r.cursor++
			}
			// case "enter", " ":
			// 	_, ok := m.selected[m.cursor]
			// 	if ok {
			// 		delete(m.selected, m.cursor)
			// 	} else {
			// 		m.selected[m.cursor] = struct{}{}
			// 	}
		}
	}
	return r, nil
}

func (r RecipientList) View() string {
	var b strings.Builder
	limit := r.width
	if max := len(r.recipients); limit*2 < max {
		limit = max
	}
	for i := r.cursor; i < r.cursor+limit; i++ {
		a := r.recipients[i]
		fmt.Fprintf(&b, "｜%s\n｜%s\n", a.Name, a.Address)
	}
	return appStyle.Render(b.String())
}
