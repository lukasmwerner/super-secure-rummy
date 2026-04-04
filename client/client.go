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
		//vp.View()

		v := tea.NewView(lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, vp.View()))
		v.AltScreen = true

		return v
	}

	s := fmt.Sprintf("Hello %s\nYour term is %s\nYour window size is %dx%d\nBackground: %s\nColor Profile: %s", hex.EncodeToString(m.PubKey.Marshal()), m.Term, m.Width, m.Height, m.Bg, m.Color_profile)

	var setBuilder []string

	for i := 1; i < 14; i++ {
		if i == 13 {
			setBuilder = append(setBuilder, card.FullCard(card.Spade, strconv.Itoa(i), m.Bg))
		} else {
			setBuilder = append(setBuilder, card.PartialCard(card.Spade, strconv.Itoa(i), m.Bg))
		}
	}

	//setBuilder[len(setBuilder)-1] = card.FullCard(card.Spade, strconv.Itoa(13), m.Bg)

	set := lipgloss.JoinVertical(
		lipgloss.Left,
		setBuilder...,
	)

	stacks := lipgloss.JoinHorizontal(lipgloss.Top, set, " ", set)

	// helpInfo := helpStyle.Render(fmt.Sprintf("\n?: help, q: exit\n"))
	// + m.Quit_text_style.Render(helpInfo)
	var handBuilder []string

	for i := 0; i < len(m.State.Hand[m.PubKey]); i++ {
		handBuilder = append(handBuilder, card.PartialCard(m.State.Hand[m.PubKey][i].Suit, m.State.Hand[m.PubKey][i].Rank, m.Bg))
	}

	currentHand := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		handBuilder...,
	)

	centerHand := lipgloss.PlaceHorizontal(m.Width, lipgloss.Center, currentHand)
	v := tea.NewView(m.Text_style.Render(s) + "\n\n" + stacks + "\n\n" + centerHand)
	v.AltScreen = true
	return v
}
