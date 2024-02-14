package main

import (
	"strings"

	"github.com/dkotik/mdsend/userinterface/bubbletea/echobox"
	"github.com/dkotik/mdsend/userinterface/bubbletea/recipientlist"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type Window struct {
	tabFocus      int
	showEchobox   bool
	echobox       *echobox.Model
	recipientList *recipientlist.Model
	progress      progress.Model
}

func (w Window) Init() tea.Cmd {
	return tea.Batch(
		w.echobox.Init(),
		w.recipientList.Init(),
	)
}

func (w Window) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.progress.Width = msg.Width
		if msg.Width < 48 {
			w.showEchobox = false
			l, cmd := w.recipientList.Update(tea.WindowSizeMsg{
				Height: msg.Height - 1,
				Width:  msg.Width,
			})
			rl := l.(recipientlist.Model)
			w.recipientList = &rl
			return w, cmd
		}
		w.showEchobox = true

		m, cmd := w.echobox.Update(tea.WindowSizeMsg{
			Height: msg.Height - 1,
			Width:  msg.Width - 24,
		})
		echobox := m.(echobox.Model)
		w.echobox = &echobox

		l, cmd2 := w.recipientList.Update(tea.WindowSizeMsg{
			Height: msg.Height - 1,
			Width:  24,
		})
		rl := l.(recipientlist.Model)
		w.recipientList = &rl

		return w, tea.Batch(cmd, cmd2)
	case progress.FrameMsg: // animate progress bar
		progressModel, cmd := w.progress.Update(msg)
		w.progress = progressModel.(progress.Model)
		return w, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return w, tea.Quit
		case "tab":
			if w.tabFocus == 0 {
				w.tabFocus = 1
			} else {
				w.tabFocus = 0
			}
			return w, nil
		default:
			if w.showEchobox && w.tabFocus > 0 {
				m, cmd := w.echobox.Update(msg)
				echobox := m.(echobox.Model)
				w.echobox = &echobox
				return w, cmd
			}

			m, cmd := w.recipientList.Update(msg)
			rl := m.(recipientlist.Model)
			w.recipientList = &rl
			return w, cmd
		}
	}

	m, cmd := w.echobox.Update(msg)
	echobox := m.(echobox.Model)
	w.echobox = &echobox

	l, cmd2 := w.recipientList.Update(msg)
	rl := l.(recipientlist.Model)
	w.recipientList = &rl

	return w, tea.Batch(cmd, cmd2)
}

func (w Window) View() string {
	var result []string
	right := w.recipientList.Render()

	if w.showEchobox {
		left := w.echobox.Render()
		// window := len(left)

		for i := 0; i < len(left); i++ {
			result = append(result, left[i]+right[i])
		}
	} else {
		for i := 0; i < len(right); i++ {
			result = append(result, right[i])
		}
	}

	result = append(result, w.progress.View())
	return strings.Join(result, "\n")
}
