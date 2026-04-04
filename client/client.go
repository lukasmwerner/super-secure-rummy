package client

import (
	"fmt"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lukasmwerner/secure-rummy/card"
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type Model struct {
	Term            string
	Width           int
	Height          int
	Bg              string
	Color_profile   string
	Text_style      lipgloss.Style
	Quit_text_style lipgloss.Style
	Help            bool
	//state           sessionState
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.ColorProfileMsg:
		m.Color_profile = msg.String()
	case tea.BackgroundColorMsg:
		if msg.IsDark() {
			m.Bg = "dark"
		} else {
			m.Bg = "light"
		}
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if !m.Help {
				return m, tea.Quit
			}
			m.Help = false
		case "?":
			m.Help = true
		}
	}
	return m, nil
}

func (m Model) View() tea.View {

	//model := m.currentFocusedModel()
	if m.Help {
		vp := viewport.New()
		vp.SetWidth(m.Width / 2)
		vp.SetHeight(m.Height / 2)
		vp.Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			PaddingRight(2)
		vp.SetContent(fmt.Sprintf("Welcome to Secure Rummy!\nHere are the basic commands:\n ...\n"))

		v := tea.NewView(lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, vp.View()))
		v.AltScreen = true

		return v
	}

	s := fmt.Sprintf("Your term is %s\nYour window size is %dx%d\nBackground: %s\nColor Profile: %s", m.Term, m.Width, m.Height, m.Bg, m.Color_profile)

	meld_cards := []*lipgloss.Layer{
		card.CardLayer(card.Spade, "K", m.Bg),
		card.CardLayer(card.Spade, "Q", m.Bg),
		card.CardLayer(card.Spade, "J", m.Bg),
	}
	meld := lipgloss.NewCompositor(
		IMap(meld_cards, func(i int, l *lipgloss.Layer) *lipgloss.Layer { return l.Y(i * 3).Z(i) })...,
	).Render()
	stacks := lipgloss.JoinHorizontal(lipgloss.Top, meld, " ", meld)

	hand_cards := []*lipgloss.Layer{
		card.CardLayer(card.Club, "A", m.Bg).Y(1),
		card.CardLayer(card.Heart, "2", m.Bg).Y(1),
		card.CardLayer(card.Diamond, "J", m.Bg).Y(0),
		card.CardLayer(card.Spade, "4", m.Bg).Y(0),
	}
	hand := lipgloss.NewCompositor(
		IMap(hand_cards, func(i int, l *lipgloss.Layer) *lipgloss.Layer { return l.X(i * 3).Z(i) })...,
	)
	leftPad := lipgloss.NewStyle().Width((m.Width - hand.Bounds().Dy()) / 2)
	centerHand := lipgloss.JoinHorizontal(lipgloss.Bottom, leftPad.Render(" "), hand.Render())
	v := tea.NewView(m.Text_style.Render(s) + "\n\n" + stacks + "\n\n" + centerHand)
	v.AltScreen = true
	return v
}

func IMap[T, V any](ts []T, fn func(int, T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(i, t)
	}
	return result
}
