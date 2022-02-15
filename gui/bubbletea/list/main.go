package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var rs []Recipient

	for i := 0; i < 100; i++ {
		rs = append(rs, Recipient{
			Name:    fmt.Sprintf("Friend #%d", i),
			Address: "test@gmail.com",
		})
	}

	p := tea.NewProgram(RecipientList{
		recipients: rs,
	})
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
