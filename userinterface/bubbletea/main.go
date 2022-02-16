package main

import (
	"fmt"
	"mdsend/userinterface/bubbletea/recipientlist"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var rs []recipientlist.Recipient

	for i := 0; i < 500; i++ {
		rs = append(rs, recipientlist.Recipient{
			Name:    fmt.Sprintf("Friend #%d", i),
			Address: "test@gmail.com",
			State:   recipientlist.DeliveryState(i) % 3,
		})
	}

	p := tea.NewProgram(recipientlist.Model{
		Recipients:    rs,
		ControlsTimer: make(chan (*struct{}), 1),
	}, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

/*
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
*/
