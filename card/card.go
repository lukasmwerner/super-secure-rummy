package card

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Suit string

const (
	width      = 8
	fullHeight = 6
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

	blackText = lipgloss.Color("#000000")
	whiteText = lipgloss.Color("#ffffff")
)

// Returns the matching color to whatever the background is set to
func adaptiveColor(bg string, light color.Color, dark color.Color) color.Color {
	if bg == "dark" {
		return dark
	}
	return light
}

func FullCard(s Suit, r string, bg string) string {
	style := lipgloss.NewStyle().Foreground(adaptiveColor(bg, blackText, whiteText))
	if s == Diamond || s == Heart {
		style = redCard
	}

	left := lipgloss.PlaceVertical(fullHeight-1, lipgloss.Top, style.Render(string(s)))
	left = lipgloss.JoinVertical(lipgloss.Left, style.Render(r), left)

	// TODO: properly render the middle design of the card
	filler := fillerStyle.Render(" ")

	right := lipgloss.PlaceVertical(fullHeight-1, lipgloss.Bottom, style.Render(string(s)))
	right = lipgloss.JoinVertical(lipgloss.Right, right, style.Render(r))

	contents := lipgloss.JoinHorizontal(lipgloss.Center, left, filler, right)
	return fullCardBorder.Render(contents)

}

func PartialCard(s Suit, r string, bg string) string {
	style := lipgloss.NewStyle().Foreground(adaptiveColor(bg, blackText, whiteText))
	if s == Diamond || s == Heart {
		style = redCard
	}

	contents := style.Render(r) + "\n" + style.Render(string(s))
	return partialCardBorder.Render(contents)
}

func RaisedCard(s Suit, r string, bg string) string {
	style := lipgloss.NewStyle().Foreground(adaptiveColor(bg, blackText, whiteText))
	if s == Diamond || s == Heart {
		style = redCard
	}
	contents := style.Render(r) + "\n" + style.Render(string(s)) + "\n\n"
	return partialCardBorder.Render(contents)

}

func DrawPile() string {
	blankCard := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(width + 2).Height(fullHeight + 2)
	return blankCard.Render(" ")
}
