package client

import (
	"fmt"
	"math/rand"
	"strconv"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lukasmwerner/super-secure-rummy/card"
	"github.com/lukasmwerner/super-secure-rummy/game"
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// Focus Target is which element (hand, discard, draw, or melds) is currently selected
type Model struct {
	PubKey          string
	Term            string
	Width           int
	Height          int
	Bg              string
	Color_profile   string
	Text_style      lipgloss.Style
	Quit_text_style lipgloss.Style
	Help            bool
	State           *game.State
	MeldID          int
	MeldLen         []int
	HandLen         int
	HandID          int
	FocusTarget     int
}

// For conversion from FocusTarget to Text
var FocusTargetMap = []string{"Hand", "Draw", "Discard", "Melds"}

func (m Model) Init() tea.Cmd {
	m.State.Hand[m.PubKey] = []*lipgloss.Layer{}
	m.HandLen = 7
	for i := 0; i < 7; i++ {
		c := card.HalfCardLayer(false, card.SuitMap[rand.Intn(4)], strconv.Itoa(rand.Intn(13)+1), m.Bg)

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
			if m.HandLen < 12 {
				m.HandLen++
				c := card.HalfCardLayer(false, card.SuitMap[rand.Intn(4)], strconv.Itoa(rand.Intn(13)+1), m.Bg)
				m.State.Hand[m.PubKey] = append(m.State.Hand[m.PubKey], c)
			}
		case "down":
			if m.HandLen > 0 {
				m.HandLen--
				m.State.Hand[m.PubKey] = m.State.Hand[m.PubKey][:len(m.State.Hand[m.PubKey])-1]
			}
		case "-":
			if m.MeldLen[m.MeldID] > 0 {
				m.MeldLen[m.MeldID]--
			}
		case "+", "=":
			if m.MeldLen[m.MeldID] < 13 {
				m.MeldLen[m.MeldID]++
			}
		case "left":
			if m.MeldID > 0 {
				m.MeldID--
			}
		case "right":
			if m.MeldID < len(m.MeldLen) {
				m.MeldID++
			}
		case "tab":
			if m.FocusTarget < 3 {
				m.FocusTarget++
			} else {
				m.FocusTarget = 0
			}
		}

	}
	return m, nil
}

func (m Model) View() tea.View {

	if m.Help {
		vp := viewport.New()
		vp.SetWidth(m.Width / 2)
		vp.SetHeight(m.Height / 2)
		vp.Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			PaddingRight(2)
		vp.SetContent(fmt.Sprintf("Welcome to Secure Rummy!\nHere are the basic commands:\n (+/-): change stack heights\n (up/down): change hand card count\n (left/right): change which stack is selected"))

		v := tea.NewView(lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, vp.View()))
		v.AltScreen = true

		return v
	}

	s := fmt.Sprintf("Hello %s\nYour term is %s\nYour window size is %dx%d\nBackground: %s\nColor Profile: %s\n m.meldlen: %d\n m.handlen %d\n m.meldid %d\n m.FocusTarget %s\n", m.PubKey, m.Term, m.Width, m.Height, m.Bg, m.Color_profile, m.MeldLen, m.HandLen, m.MeldID, FocusTargetMap[m.FocusTarget])

	var meldBuilder [][]*lipgloss.Layer

	var meld []string

	for j := 0; j < len(m.MeldLen); j++ {
		meldBuilder = append(meldBuilder, []*lipgloss.Layer{})
		for i := 1; i < m.MeldLen[j]+1; i++ {
			if i == m.MeldLen[j] {
				meldBuilder[j] = append(meldBuilder[j], card.CardLayer(card.SuitMap[j%4], strconv.Itoa(i), m.Bg))
			} else {
				meldBuilder[j] = append(meldBuilder[j], card.HalfCardLayer(false, card.SuitMap[j%4], strconv.Itoa(i), m.Bg))
			}
		}

		meld = append(meld, lipgloss.NewCompositor(
			IMap(meldBuilder[j], func(i int, l *lipgloss.Layer) *lipgloss.Layer { return l.Y(i * 3).Z(i) })...,
		).Render())

	}

	stacks := lipgloss.JoinHorizontal(lipgloss.Top, meld...)

	hand := lipgloss.NewCompositor(
		IMap(m.State.Hand[m.PubKey], func(i int, l *lipgloss.Layer) *lipgloss.Layer { return l.X(i * 3).Z(i) })...,
	)
	handRender := hand.Render()
	meldsRender := stacks
	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	if m.FocusTarget == 0 { // Hand
		handRender = borderStyle.Render(hand.Render())
	} else if m.FocusTarget == 1 { // Draw

	} else if m.FocusTarget == 2 { // Discard

	} else if m.FocusTarget == 3 { // Melds
		meldsRender = borderStyle.Render(stacks)
	}

	leftPad := lipgloss.NewStyle().Width((m.Width - hand.Bounds().Dx()) / 2)

	centerHand := lipgloss.JoinHorizontal(lipgloss.Bottom, leftPad.Render(" "), handRender)

	v := tea.NewView(m.Text_style.Render(s) + "\n\n" + meldsRender +
		"\n\n" + centerHand)
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
