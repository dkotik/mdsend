package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#43884a")).
			Background(lipgloss.Color("#3acf49"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff77a8")).
			Background(lipgloss.Color("#d64ce7"))
)

type (
	DeliveryState uint8
	// TickMsg         time.Time
	HideControlsMsg time.Time
)

const (
	DeliveryStatePending DeliveryState = iota
	DeliveryStateSent
	DeliveryStateFailed

	delay = time.Millisecond * 1000
)

type Recipient struct {
	Name    string
	Address string
	State   DeliveryState
	// Fields  map[string]interface{}
}

type RecipientList struct {
	cursor                int
	cursorMaximum         int
	scrollMomentum        int
	scrollMomentumMaximum int
	scrollMomentumTime    time.Time
	width                 int
	height                int
	recipients            []Recipient
	controlsTimer         chan (*struct{})
	showControls          bool
}

func (r *RecipientList) triggerControlsActivation() tea.Cmd {
	r.showControls = true
	return func() tea.Msg {
		select { // attempt to send a message to the channel
		case r.controlsTimer <- &struct{}{}:
			return nil
		default:
			return nil
		}
	}
}

func (r *RecipientList) listenForControlsActivation() tea.Cmd {
	return func() tea.Msg {
		msg := <-r.controlsTimer
		if msg == nil { // channel closed
			return nil
		}

		timer := time.NewTimer(delay)
		for {
			select {
			case msg := <-r.controlsTimer:
				if msg == nil { // channel closed
					return nil
				}
				if !timer.Stop() {
					<-timer.C // drain channel, see docs
				}
				timer.Reset(delay)
			case t := <-timer.C:
				return HideControlsMsg(t)
			}
		}
	}
}

func (r RecipientList) Init() tea.Cmd {
	// r.controlsTimer = make(chan (*struct{}), 1) // TODO: replace with external channel that closes? or add a closer interface
	// r.scrollMomentumMaximum = len(r.recipients)/5 + 30
	cmd := r.listenForControlsActivation()
	return cmd
}

func (r *RecipientList) getMomentum() int {
	t := time.Now()
	if t.After(r.scrollMomentumTime) {
		r.scrollMomentum = 1 // reset
	} else if r.scrollMomentum < r.scrollMomentumMaximum {
		r.scrollMomentum += r.scrollMomentumMaximum / 3
	}
	r.scrollMomentumTime = t.Add(delay)
	return r.scrollMomentum
}

func (r RecipientList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width, r.height = msg.Width, msg.Height
		r.cursorMaximum = len(r.recipients)*2 - r.height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return r, tea.Quit
		case "up", "k":
			r.cursor -= r.getMomentum()
			if r.cursor < 0 {
				r.cursor = 0
			}
			cmd := r.triggerControlsActivation()
			return r, cmd
		case "down", "j":
			r.cursor += r.getMomentum()
			if r.cursor > r.cursorMaximum {
				r.cursor = r.cursorMaximum
			}
			// cmd := tea.Tick(delay, func(t time.Time) tea.Msg {
			// 	return TickMsg(t)
			// })
			cmd := r.triggerControlsActivation()
			return r, cmd
		}
	case HideControlsMsg:
		r.showControls = false
		cmd := r.listenForControlsActivation()
		return r, cmd
	}
	return r, nil
}

func (r RecipientList) View() string {
	result := make([]string, r.height)
	print := func(i int, s string) {
		if len(s) > r.width {
			result[i] = s[:r.width]
			return
		}
		result[i] = fmt.Sprintf("%*s", -r.width, s) // pad
		// result[i] = strings.Replace(fmt.Sprintf("%*s", -r.width, s), " ", "~", -1) // pad
	}

	for i := 0; i < r.height; i++ {
		offset := r.cursor + i
		index := offset / 2
		if index > len(r.recipients)-1 {
			print(i, fmt.Sprintf("│end %d/%d            ", i+1, r.height))
			continue
		}

		rs := r.recipients[index]
		if offset%2 == 0 {
			state := '?'
			switch rs.State {
			case DeliveryStateSent:
				state = '>'
			case DeliveryStateFailed:
				state = '!'
			}

			print(i, fmt.Sprintf("│%d/%d %c %s", i+1, r.height, state, rs.Name))
		} else {
			print(i, fmt.Sprintf("│%d/%d %s", i+1, r.height, rs.Address))
		}
	}

	// if time.Now().Before(r.scrollMomentumTime) {
	if r.showControls {
		cutoff := r.width + 1
		scrolled := int(float32(r.cursor)/float32(r.cursorMaximum)*float32(r.height)) - 1
		for i := 0; i < r.height; i++ {
			if i == scrolled || i == 0 && scrolled < 1 {
				result[i] = result[i][:cutoff] + appStyle.Render(`-`) // `│`
			} else {
				result[i] = result[i][:cutoff] + helpStyle.Render(`:`)
			}
		}
	}

	return strings.Join(result, "\n")
}
