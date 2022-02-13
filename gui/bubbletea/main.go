package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
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

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	var messages = make(chan (msgDelivered))
	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)+50))
			messages <- msgDelivered{
				Address:          "boo@boo.com",
				ConfirmationCode: fmt.Sprintf("%d", rand.Intn(99999999)),
			}
		}
	}()

	p := tea.NewProgram(model{
		progress: progress.New(
			progress.WithDefaultGradient(),
			progress.WithoutPercentage(),
		),
		messages: messages,
	},
		tea.WithAltScreen(),
	)

	// TODO: add daemon mode (see examples in bubbletea)
	if p.Start() != nil {
		fmt.Println("could not start program")
		os.Exit(1)
	}
}
