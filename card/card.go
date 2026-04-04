package card

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

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
	partialCardBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true, true, false, true).
				Width(width + 2)
	fullCardBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())
	fillerStyle = lipgloss.NewStyle().
			Width(width - 2)
	redCard = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#931B29"))
)

// Returns the matching color to whatever the background is set to
func adaptiveColor(bg string, light color.Color, dark color.Color) color.Color {
	if bg == "dark" {
		return dark
	}
	return light
}

func FullCard(s Suit, c Color, r string, bg string) string {
	style := lipgloss.NewStyle().Foreground(adaptiveColor(bg, lipgloss.Black, lipgloss.White))
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
	return fullCardBorder.Render(contents)

}

func PartialCard(s Suit, c Color, r string, bg string) string {
	style := lipgloss.NewStyle().Foreground(adaptiveColor(bg, lipgloss.Black, lipgloss.White))
	if c == Red {
		style = redCard
	}

	contents := style.Render(r) + "\n" + style.Render(string(s))
	return partialCardBorder.Render(contents)
}
