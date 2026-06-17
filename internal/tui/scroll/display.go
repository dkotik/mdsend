package scroll

import "github.com/charmbracelet/lipgloss"

var (
	DisplayBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#43884a")).
			Background(lipgloss.Color("#3acf49")).
			Render(`-`) // `│`

	DisplayTrack = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff77a8")).
			Background(lipgloss.Color("#403441")).
			Render(`:`)
)
