/*
Package tealist provides a [bubbletea.Model] for [list.List].
*/
package tealist

import (
	"github.com/dkotik/mdsend/list"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO: check out https://github.com/charmbracelet/huh forms
// that should be able to do all the basics for input
type List struct {
	list.List
}

func (l *List) Init() tea.Cmd {
	return nil
}

func (l *List) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return l, nil
}

func (l *List) View() string {
	return "list"
}
