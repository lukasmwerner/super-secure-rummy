package client

import (
	"fmt"
	"strings"

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

	var st strings.Builder
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

	s := fmt.Sprintf("Your term is %s\nYour window size is %dx%d\nBackground: %s\nColor Profile: %s", m.Term, m.Width, m.Height, m.Bg, m.Color_profile)
	st.WriteString(helpStyle.Render(fmt.Sprintf("\n?: help, q: exit\n")))

	set := lipgloss.JoinVertical(
		lipgloss.Left,
		card.PartialCard(card.Spade, card.Black, "K", m.Bg),
		card.PartialCard(card.Spade, card.Red, "Q", m.Bg),
		card.FullCard(card.Spade, card.Black, "J", m.Bg),
	)

	stacks := lipgloss.JoinHorizontal(lipgloss.Top, set, " ", set)

	v := tea.NewView(m.Text_style.Render(s) + "\n\n" + stacks + "\n\n" + m.Quit_text_style.Render(st.String()))
	v.AltScreen = true
	return v
}
