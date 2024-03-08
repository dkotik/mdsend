package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Line struct {
	Value string
}

func (l *Line) Init() tea.Cmd {
	return nil
}

func (l *Line) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return l, tea.Quit
		default:
			l.Value += msg.String()
		}
	}
	return l, nil
}

func (l *Line) View() string {
	return time.Now().Format(time.RFC3339Nano)
	// return l.Value
}

func main() {
	p := tea.NewProgram(&Line{})
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
