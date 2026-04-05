package game

import "math/rand/v2"

// Deck represents a stack of cards that can be drawn from.
type Deck struct {
	cards []Card
}

// NewDeck creates a standard 52-card deck (no jokers), unshuffled.
// Cards are ordered by suit (Spades, Hearts, Diamonds, Clubs) and rank (A through K).
func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for suit := Spades; suit <= Clubs; suit++ {
		for rank := 1; rank <= 13; rank++ {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}
	return &Deck{cards: cards}
}

// Shuffle randomizes the order of cards in the deck using Fisher-Yates.
func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

// Draw removes and returns the top card from the deck.
// Returns false if the deck is empty.
func (d *Deck) Draw() (Card, bool) {
	if len(d.cards) == 0 {
		return Card{}, false
	}
	card := d.cards[len(d.cards)-1]
	d.cards = d.cards[:len(d.cards)-1]
	return card, true
}

// Remaining returns the number of cards left in the deck.
func (d *Deck) Remaining() int {
	return len(d.cards)
}

// IsEmpty returns true if no cards remain.
func (d *Deck) IsEmpty() bool {
	return len(d.cards) == 0
}
