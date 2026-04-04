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

var SuitMap = []Suit{Diamond, Heart, Spade, Club}

type Card struct {
	Suit Suit
	Rank string
}

var (
	halfCardBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			Width(width + 2)
	fullCardBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())
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

func cardDesign(s Suit, rank string, bg string) string {
	style := lipgloss.NewStyle().Foreground(adaptiveColor(bg, blackText, whiteText))
	if s == Diamond || s == Heart {
		style = redCard
	}

	RankConv(&rank)

	left := lipgloss.PlaceVertical(fullHeight-1, lipgloss.Top, style.Render(string(s)))
	left = lipgloss.JoinVertical(lipgloss.Left, style.Render(rank), left)

	// Subtract 2 for each char in the rank because of the bottom right
	filler := lipgloss.NewStyle().Width(width - (2 * len(string(rank)))).Render("")

	right := lipgloss.PlaceVertical(fullHeight-1, lipgloss.Bottom, style.Render(string(s)))
	right = lipgloss.JoinVertical(lipgloss.Right, right, style.Render(rank))

	contents := lipgloss.JoinHorizontal(lipgloss.Center, left, filler, right)
	return fullCardBorder.Render(contents)
}

func halfCardDesign(active bool, s Suit, rank string, bg string) string {
	style := lipgloss.NewStyle().Foreground(adaptiveColor(bg, blackText, whiteText))
	if s == Diamond || s == Heart {
		style = redCard
	}

	height := (fullHeight / 2) - 1
	if active {
		height += 1
	}

	RankConv(&rank)

	left := lipgloss.PlaceVertical(height, lipgloss.Top, style.Render(string(s)))
	left = lipgloss.JoinVertical(lipgloss.Left, style.Render(rank), left)

	filler := lipgloss.NewStyle().Width(width - 2 - len(string(rank))).Render("")

	contents := lipgloss.JoinHorizontal(lipgloss.Center, left, filler, " ")
	return halfCardBorder.Render(contents)
}

func CardLayer(s Suit, rank string, bg string) *lipgloss.Layer {
	return lipgloss.NewLayer(cardDesign(s, rank, bg))
}

func HalfCardLayer(active bool, s Suit, rank string, bg string) *lipgloss.Layer {
	return lipgloss.NewLayer(halfCardDesign(active, s, rank, bg))
}

func DrawPile() *lipgloss.Layer {
	blankCard := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(width + 2).Height(fullHeight + 2)
	return lipgloss.NewLayer(blankCard.Render(""))
}

func RankConv(r *string) {
	switch *r {
	case "1":
		*r = "A"
	case "11":
		*r = "J"
	case "12":
		*r = "Q"
	case "13":
		*r = "K"
	}
}
