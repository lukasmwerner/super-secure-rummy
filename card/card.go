package card

import "charm.land/lipgloss/v2"

type Color int
type Suit string

const (
	width      = 8
	fullHeight = 6
)
const (
	Black Color = iota
	Red
)
const (
	Diamond Suit = "♦︎"
	Spade   Suit = "♠︎"
	Club    Suit = "♣︎"
	Heart   Suit = "♥︎"
)

var (
	cardBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())
	fillerStyle = lipgloss.NewStyle().
			Width(width - 2)
	redCard = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#931B29"))
)

func FullCard(s Suit, c Color, r string) string {
	style := lipgloss.NewStyle()
	if c == Red {
		style = redCard
	}

	left := lipgloss.PlaceVertical(fullHeight-1, lipgloss.Top, style.Render(r))
	left = lipgloss.JoinVertical(lipgloss.Left, style.Render(string(s)), left)

	// TODO: properly render the middle design of the card
	filler := fillerStyle.Render(" ")

	right := lipgloss.PlaceVertical(fullHeight-1, lipgloss.Bottom, style.Render(string(s)))
	right = lipgloss.JoinVertical(lipgloss.Right, right, style.Render(r))

	contents := lipgloss.JoinHorizontal(lipgloss.Center, left, filler, right)
	return cardBorder.Render(contents)

}
