package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/userinterface/bubbletea/echobox"
	"github.com/dkotik/mdsend/userinterface/bubbletea/recipientlist"
	"github.com/dkotik/mdsend/userinterface/bubbletea/scroll"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#aa2f1c"))
	// Background(lipgloss.Color("#33c987"))

	var rs []recipientlist.Recipient

	for i := 0; i < 500; i++ {
		rs = append(rs, recipientlist.Recipient{
			Name:    fmt.Sprintf("Friend #%d", i),
			Address: "test@gmail.com",
			State:   recipientlist.DeliveryState(i) % 3,
		})
	}

	progressBar := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	progressBar.Empty = '█'
	progressBar.EmptyColor = "#3a85aa"

	p := tea.NewProgram(
		Window{
			echobox: &echobox.Model{
				Momentum: scroll.NewMomentum(3, time.Millisecond*10),
			},
			recipientList: &recipientlist.Model{
				Recipients:    rs,
				ControlsTimer: make(chan (*struct{}), 1),
				Momentum:      scroll.NewMomentum(5, time.Millisecond*10),
			},
			progress: progressBar,
		}, tea.WithAltScreen())

	go func() {
		for i := 0; i < 100; i++ {
			p.Send(echobox.EchoMsg{
				Message: fmt.Sprintf("boo %s %d", time.Now(), i+1),
				Style:   style,
			})
			time.Sleep(time.Millisecond * 2)
			p.Send(scroll.ToBottom)
		}
	}()

	if err := p.Start(); err != nil {
		fmt.Printf("mdsend %s", mdsend.Version)
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
