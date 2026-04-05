package game

import "sort"

// MeldType distinguishes sets from runs.
type MeldType int

const (
	MeldSet MeldType = iota // 3-4 cards of same rank, different suits
	MeldRun                 // 3+ consecutive cards of same suit
)

// Meld represents a group of cards laid on the table.
type Meld struct {
	Type  MeldType
	Cards []Card
}

// ValidateSet checks if cards form a valid set: 3-4 cards of the same rank,
// all with different suits. Returns true if valid.
func ValidateSet(cards []Card) bool {
	if len(cards) < 3 || len(cards) > 4 {
		return false
	}

	rank := cards[0].Rank
	suits := make(map[Suit]bool)

	for _, c := range cards {
		if c.Rank != rank {
			return false
		}
		if suits[c.Suit] {
			return false // duplicate suit
		}
		suits[c.Suit] = true
	}
	return true
}

// ValidateRun checks if cards form a valid run: 3+ consecutive cards
// of the same suit. Ace is low only (A-2-3 is valid, Q-K-A is not).
// Cards do not need to be pre-sorted.
func ValidateRun(cards []Card) bool {
	if len(cards) < 3 {
		return false
	}

	suit := cards[0].Suit

	// Sort by rank
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rank < sorted[j].Rank
	})

	for i, c := range sorted {
		if c.Suit != suit {
			return false
		}
		if i > 0 && c.Rank != sorted[i-1].Rank+1 {
			return false
		}
	}
	return true
}

// ValidateMeld checks if cards form either a valid set or run.
// Returns the MeldType and whether the meld is valid.
func ValidateMeld(cards []Card) (MeldType, bool) {
	if ValidateSet(cards) {
		return MeldSet, true
	}
	if ValidateRun(cards) {
		return MeldRun, true
	}
	return 0, false
}

// CanLayOff checks if a card can be legally added to an existing meld.
// For sets: card must have the same rank and a suit not already in the meld (max 4).
// For runs: card must extend the sequence at either end with the same suit.
func CanLayOff(card Card, meld Meld) bool {
	switch meld.Type {
	case MeldSet:
		if len(meld.Cards) >= 4 {
			return false
		}
		if card.Rank != meld.Cards[0].Rank {
			return false
		}
		for _, c := range meld.Cards {
			if c.Suit == card.Suit {
				return false
			}
		}
		return true

	case MeldRun:
		if card.Suit != meld.Cards[0].Suit {
			return false
		}
		// Find min and max rank in the run
		minRank, maxRank := meld.Cards[0].Rank, meld.Cards[0].Rank
		for _, c := range meld.Cards {
			if c.Rank < minRank {
				minRank = c.Rank
			}
			if c.Rank > maxRank {
				maxRank = c.Rank
			}
		}
		// Card must extend at either end
		return card.Rank == minRank-1 || card.Rank == maxRank+1
	}
	return false
}
