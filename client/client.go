package client

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/ssh"
	"github.com/lukasmwerner/secure-rummy/card"
	"github.com/lukasmwerner/secure-rummy/game"
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type Model struct {
	PubKey          ssh.PublicKey
	Term            string
	Width           int
	Height          int
	Bg              string
	Color_profile   string
	Text_style      lipgloss.Style
	Quit_text_style lipgloss.Style
	Help            bool
	State           game.State
	Temp            int
	//state           sessionState
}

func (m Model) Init() tea.Cmd {
	m.State.Hand[m.PubKey] = []card.Card{}
	// init game state with data
	for i := 0; i < 7; i++ {

		c := card.Card{
			Suit: card.SuitMap[rand.Intn(4)],
			Rank: strconv.Itoa(rand.Intn(13) + 1),
		}

		m.State.Hand[m.PubKey] = append(m.State.Hand[m.PubKey], c)
	}

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
		case "up":
			c := card.Card{
				Suit: card.SuitMap[rand.Intn(4)],
				Rank: strconv.Itoa(rand.Intn(13) + 1),
			}
			m.State.Hand[m.PubKey] = append(m.State.Hand[m.PubKey], c)
		case "down":
			if len(m.State.Hand[m.PubKey]) > 0 {
				m.State.Hand[m.PubKey] = m.State.Hand[m.PubKey][:len(m.State.Hand[m.PubKey])-1]
			}
		case "-":
			if m.Temp > 0 {
				m.Temp--
			}
		case "+", "=":
			if m.Temp < 13 {
				m.Temp++
			}
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
		vp.SetContent(fmt.Sprintf("Welcome to Secure Rummy!\nHere are the basic commands:\n (+/-): change stack heights\n (up/down): change hand card count\n"))

		v := tea.NewView(lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, vp.View()))
		v.AltScreen = true

		return v
	}

	s := fmt.Sprintf("Hello %s\nYour term is %s\nYour window size is %dx%d\nBackground: %s\nColor Profile: %s\n m.temp: %d\n", hex.EncodeToString(m.PubKey.Marshal()), m.Term, m.Width, m.Height, m.Bg, m.Color_profile, m.Temp)

// <<<<<<< sessions
// 	var setBuilder [][]string
// 	var set = []string{}
// 	for j := 0; j < 4; j++ {
// 		setBuilder = append(setBuilder, []string{})
// 		for i := 1; i < m.Temp+1; i++ {
// 			if i == m.Temp {
// 				setBuilder[j] = append(setBuilder[j], card.FullCard(card.SuitMap[j], strconv.Itoa(i), m.Bg))
// 			} else {
// 				setBuilder[j] = append(setBuilder[j], card.PartialCard(card.SuitMap[j], strconv.Itoa(i), m.Bg))
// 			}
// 		}

// 		set = append(set, lipgloss.JoinVertical(
// 			lipgloss.Left,
// 			setBuilder[j]...,
// 		))
// 	}

// 	stacks := lipgloss.JoinHorizontal(lipgloss.Top, set...)

// 	// helpInfo := helpStyle.Render(fmt.Sprintf("\n?: help, q: exit\n"))
// 	// + m.Quit_text_style.Render(helpInfo)
// 	var handBuilder []string

// 	for i := 0; i < len(m.State.Hand[m.PubKey]); i++ {
// 		handBuilder = append(handBuilder, card.PartialCard(m.State.Hand[m.PubKey][i].Suit, m.State.Hand[m.PubKey][i].Rank, m.Bg))
// 	}

// 	currentHand := lipgloss.JoinHorizontal(
// 		lipgloss.Bottom,
// 		handBuilder...,
// 	)

// 	centerHand := lipgloss.PlaceHorizontal(m.Width, lipgloss.Center, currentHand)
// =======
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
		card.HalfCardLayer(false, card.Club, "A", m.Bg).Y(1),
		card.HalfCardLayer(true, card.Heart, "2", m.Bg).Y(0),
		card.HalfCardLayer(false, card.Diamond, "J", m.Bg).Y(1),
		card.HalfCardLayer(false, card.Spade, "4", m.Bg).Y(1),
	}
	hand := lipgloss.NewCompositor(
		IMap(hand_cards, func(i int, l *lipgloss.Layer) *lipgloss.Layer { return l.X(i * 3).Z(i) })...,
	)
	leftPad := lipgloss.NewStyle().Width((m.Width - hand.Bounds().Dy()) / 2)
	centerHand := lipgloss.JoinHorizontal(lipgloss.Bottom, leftPad.Render(" "), hand.Render())
// >>>>>>> main
  
  
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
