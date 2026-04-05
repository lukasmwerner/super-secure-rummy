package game

import "fmt"

// Suit represents a card suit.
type Suit int

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

var suitNames = [...]string{"♠", "♥", "♦", "♣"}

func (s Suit) String() string {
	if s < Spades || s > Clubs {
		return "?"
	}
	return suitNames[s]
}

// Card represents a single playing card with a suit and rank.
type Card struct {
	Suit Suit
	Rank int // 1=Ace, 2-10, 11=Jack, 12=Queen, 13=King
}

var rankNames = [...]string{"", "A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

// String returns a human-readable card representation like "A♠" or "10♥".
func (c Card) String() string {
	if c.Rank < 1 || c.Rank > 13 {
		return fmt.Sprintf("?%s", c.Suit)
	}
	return rankNames[c.Rank] + c.Suit.String()
}

// Points returns the deadwood point value of this card.
// Ace=1, Face cards (J/Q/K)=10, others=face value.
func (c Card) Points() int {
	if c.Rank >= 10 {
		return 10
	}
	return c.Rank
}

// Equal returns true if two cards have the same suit and rank.
func (c Card) Equal(other Card) bool {
	return c.Suit == other.Suit && c.Rank == other.Rank
}

// IsRed returns true if the card is Hearts or Diamonds.
func (c Card) IsRed() bool {
	return c.Suit == Hearts || c.Suit == Diamonds
}
