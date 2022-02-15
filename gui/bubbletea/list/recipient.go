package main

import (
	"fmt"
	"strings"
	"time"

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
	cursor             int
	scrollMomentum     int
	scrollMomentumTime time.Time
	width              int
	height             int
	recipients         []Recipient
}

func (r RecipientList) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (r RecipientList) getMomentum(up bool) int {
	t := time.Now()
	if t.After(r.scrollMomentumTime) {
		r.scrollMomentum = 0 // reset
		r.scrollMomentumTime = t.Add(time.Millisecond * 300)
		if up {
			return 1
		}
		return -1
	}
	if up && r.scrollMomentum <= 1000 {
		r.scrollMomentum += 300
	} else if r.scrollMomentum >= -1000 {
		r.scrollMomentum -= 300
	}
	r.scrollMomentumTime = t.Add(time.Millisecond * 300)
	return r.scrollMomentum
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
				r.cursor -= r.getMomentum(true)
			}
		case "down", "j":
			if r.cursor < len(r.recipients)*2-r.height {
				r.cursor -= r.getMomentum(false)
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

	for i := 0; i < r.height; i++ {
		offset := r.cursor + i
		index := offset / 2
		if index >= len(r.recipients) {
			fmt.Fprintf(&b, "\n｜end %d/%d            ", i+1, r.height)
			continue
		}

		if offset%2 == 0 {
			fmt.Fprintf(&b, "\n│%d/%d %s", i+1, r.height, r.recipients[index].Name)
		} else {
			fmt.Fprintf(&b, "\n│%d/%d %s", i+1, r.height, r.recipients[index].Address)
		}
	}
	return appStyle.Render(b.String())
}
