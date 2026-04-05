package game

import (
	"errors"
	"fmt"
)

// TurnPhase tracks where a player is in their turn.
type TurnPhase int

const (
	PhaseDraw    TurnPhase = iota // Must draw from deck or discard
	PhaseAction                   // May meld or lay off (optional, repeatable)
	PhaseDiscard                  // Must discard one card to end turn
)

// cardsPerPlayer returns how many cards each player gets dealt.
func cardsPerPlayer(numPlayers int) int {
	// Standard Rummy: 7 cards for 2-4 players
	return 7
}

// Errors returned by Apply for illegal moves.
var (
	ErrNotYourTurn    = errors.New("not your turn")
	ErrWrongPhase     = errors.New("wrong phase for this action")
	ErrDeckEmpty      = errors.New("deck is empty")
	ErrDiscardEmpty   = errors.New("discard pile is empty")
	ErrInvalidMeld    = errors.New("cards do not form a valid meld")
	ErrInvalidLayOff  = errors.New("card cannot be laid off on that meld")
	ErrCardNotInHand  = errors.New("you don't have that card")
	ErrInvalidMeldIdx = errors.New("invalid meld index")
	ErrGameOver       = errors.New("game is already over")
	ErrGameNotStarted = errors.New("game has not started")
)

// PlayerState holds a single player's mutable game data.
type PlayerState struct {
	ID        string
	Name      string
	Hand      []Card
	Score     int
	Connected bool
}

// GameState is the internal, mutable, authoritative game state.
// It is only accessed by the Hub goroutine — never shared directly.
type GameState struct {
	Deck        *Deck
	DiscardPile []Card
	Players     []PlayerState
	Melds       []Meld
	CurrentTurn int // Index into Players
	Phase       TurnPhase
	GameOver    bool
	Winner      int // Index into Players, -1 if no winner
	Started     bool
}

// NewGameState creates a fresh game state for the given players.
// It creates the deck, shuffles, and deals cards.
func NewGameState(playerIDs []string) *GameState {
	gs := &GameState{
		Deck:        NewDeck(),
		DiscardPile: make([]Card, 0),
		Players:     make([]PlayerState, len(playerIDs)),
		Melds:       make([]Meld, 0),
		CurrentTurn: 0,
		Phase:       PhaseDraw,
		Winner:      -1,
		Started:     true,
	}

	for i, id := range playerIDs {
		gs.Players[i] = PlayerState{
			ID:        id,
			Name:      id,
			Hand:      make([]Card, 0),
			Score:     0,
			Connected: true,
		}
	}

	gs.Deck.Shuffle()

	// Deal cards
	numCards := cardsPerPlayer(len(playerIDs))
	for round := 0; round < numCards; round++ {
		for i := range gs.Players {
			card, ok := gs.Deck.Draw()
			if !ok {
				break // Should never happen with a standard deck and <=4 players
			}
			gs.Players[i].Hand = append(gs.Players[i].Hand, card)
		}
	}

	// Turn over the first card to start the discard pile
	if first, ok := gs.Deck.Draw(); ok {
		gs.DiscardPile = append(gs.DiscardPile, first)
	}

	return gs
}

// Apply validates and applies a GameAction, mutating the game state.
// Returns an error if the action is illegal.
func (gs *GameState) Apply(action GameAction) error {
	if gs.GameOver {
		return ErrGameOver
	}
	if !gs.Started {
		return ErrGameNotStarted
	}

	playerIdx := gs.findPlayer(action.PlayerID)
	if playerIdx == -1 {
		return fmt.Errorf("unknown player: %s", action.PlayerID)
	}
	if playerIdx != gs.CurrentTurn {
		return ErrNotYourTurn
	}

	player := &gs.Players[playerIdx]

	switch action.Type {
	case ActionDrawFromDeck:
		return gs.applyDrawFromDeck(player)
	case ActionDrawFromDiscard:
		return gs.applyDrawFromDiscard(player)
	case ActionDiscard:
		return gs.applyDiscard(player, action)
	case ActionMeld:
		return gs.applyMeld(player, action)
	case ActionLayOff:
		return gs.applyLayOff(player, action)
	default:
		return fmt.Errorf("unknown action type: %d", action.Type)
	}
}

func (gs *GameState) applyDrawFromDeck(player *PlayerState) error {
	if gs.Phase != PhaseDraw {
		return ErrWrongPhase
	}
	card, ok := gs.Deck.Draw()
	if !ok {
		// If deck is empty, reshuffle discard pile (leave top card)
		if len(gs.DiscardPile) <= 1 {
			return ErrDeckEmpty
		}
		gs.reshuffleDiscard()
		card, ok = gs.Deck.Draw()
		if !ok {
			return ErrDeckEmpty
		}
	}
	player.Hand = append(player.Hand, card)
	gs.Phase = PhaseAction
	return nil
}

func (gs *GameState) applyDrawFromDiscard(player *PlayerState) error {
	if gs.Phase != PhaseDraw {
		return ErrWrongPhase
	}
	if len(gs.DiscardPile) == 0 {
		return ErrDiscardEmpty
	}
	// Take the top card from the discard pile
	card := gs.DiscardPile[len(gs.DiscardPile)-1]
	gs.DiscardPile = gs.DiscardPile[:len(gs.DiscardPile)-1]
	player.Hand = append(player.Hand, card)
	gs.Phase = PhaseAction
	return nil
}

func (gs *GameState) applyDiscard(player *PlayerState, action GameAction) error {
	if gs.Phase != PhaseAction && gs.Phase != PhaseDiscard {
		return ErrWrongPhase
	}
	if len(action.Cards) != 1 {
		return fmt.Errorf("must discard exactly one card")
	}

	cardIdx := gs.findCardInHand(player, action.Cards[0])
	if cardIdx == -1 {
		return ErrCardNotInHand
	}

	// Remove from hand
	card := player.Hand[cardIdx]
	player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)

	// Add to discard pile
	gs.DiscardPile = append(gs.DiscardPile, card)

	// Check win condition: hand is now empty
	if len(player.Hand) == 0 {
		gs.endGame(gs.findPlayer(player.ID))
		return nil
	}

	// Advance to next player
	gs.advanceTurn()
	return nil
}

func (gs *GameState) applyMeld(player *PlayerState, action GameAction) error {
	if gs.Phase != PhaseAction {
		return ErrWrongPhase
	}
	if len(action.Cards) < 3 {
		return ErrInvalidMeld
	}

	// Validate the meld
	meldType, valid := ValidateMeld(action.Cards)
	if !valid {
		return ErrInvalidMeld
	}

	// Verify player has all the cards
	for _, c := range action.Cards {
		if gs.findCardInHand(player, c) == -1 {
			return ErrCardNotInHand
		}
	}

	// Remove cards from hand
	for _, c := range action.Cards {
		idx := gs.findCardInHand(player, c)
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}

	// Add meld to table
	gs.Melds = append(gs.Melds, Meld{Type: meldType, Cards: action.Cards})

	// Check win condition: hand is now empty (no discard needed)
	if len(player.Hand) == 0 {
		gs.endGame(gs.findPlayer(player.ID))
	}

	return nil
}

func (gs *GameState) applyLayOff(player *PlayerState, action GameAction) error {
	if gs.Phase != PhaseAction {
		return ErrWrongPhase
	}
	if len(action.Cards) != 1 {
		return fmt.Errorf("must lay off exactly one card at a time")
	}
	if action.TargetMeld < 0 || action.TargetMeld >= len(gs.Melds) {
		return ErrInvalidMeldIdx
	}

	card := action.Cards[0]
	cardIdx := gs.findCardInHand(player, card)
	if cardIdx == -1 {
		return ErrCardNotInHand
	}

	if !CanLayOff(card, gs.Melds[action.TargetMeld]) {
		return ErrInvalidLayOff
	}

	// Remove from hand
	player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)

	// Add to the meld
	gs.Melds[action.TargetMeld].Cards = append(gs.Melds[action.TargetMeld].Cards, card)

	// Check win condition
	if len(player.Hand) == 0 {
		gs.endGame(gs.findPlayer(player.ID))
	}

	return nil
}

// SnapshotFor creates a personalized, read-only view of the game for a specific player.
func (gs *GameState) SnapshotFor(playerID string) GameSnapshot {
	snap := GameSnapshot{
		DeckRemaining: gs.Deck.Remaining(),
		DiscardCount:  len(gs.DiscardPile),
		Melds:         make([]Meld, len(gs.Melds)),
		Players:       make([]PlayerInfo, len(gs.Players)),
		Phase:         gs.Phase,
		GameOver:      gs.GameOver,
		Scores:        make(map[string]int),
		Started:       gs.Started,
	}

	// Copy melds (deep copy cards)
	for i, m := range gs.Melds {
		cards := make([]Card, len(m.Cards))
		copy(cards, m.Cards)
		snap.Melds[i] = Meld{Type: m.Type, Cards: cards}
	}

	// Discard pile top card
	if len(gs.DiscardPile) > 0 {
		top := gs.DiscardPile[len(gs.DiscardPile)-1]
		snap.DiscardTop = &top
	}

	// Current turn
	if gs.CurrentTurn >= 0 && gs.CurrentTurn < len(gs.Players) {
		snap.CurrentTurn = gs.Players[gs.CurrentTurn].ID
	}

	// Winner
	if gs.GameOver && gs.Winner >= 0 {
		snap.Winner = gs.Players[gs.Winner].ID
	}

	// Player info
	for i, p := range gs.Players {
		snap.Players[i] = PlayerInfo{
			ID:        p.ID,
			Name:      p.Name,
			CardCount: len(p.Hand),
			Score:     p.Score,
			Connected: p.Connected,
			IsYou:     p.ID == playerID,
		}
		snap.Scores[p.ID] = p.Score

		// Only include the requesting player's hand
		if p.ID == playerID {
			hand := make([]Card, len(p.Hand))
			copy(hand, p.Hand)
			snap.YourHand = hand
			snap.YourTurn = (i == gs.CurrentTurn)
		}
	}

	return snap
}

// --- Internal helpers ---

func (gs *GameState) findPlayer(id string) int {
	for i, p := range gs.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func (gs *GameState) findCardInHand(player *PlayerState, card Card) int {
	for i, c := range player.Hand {
		if c.Equal(card) {
			return i
		}
	}
	return -1
}

func (gs *GameState) advanceTurn() {
	gs.CurrentTurn = (gs.CurrentTurn + 1) % len(gs.Players)
	gs.Phase = PhaseDraw

	// Skip disconnected players (with safety limit to avoid infinite loop)
	for attempts := 0; attempts < len(gs.Players); attempts++ {
		if gs.Players[gs.CurrentTurn].Connected {
			return
		}
		gs.CurrentTurn = (gs.CurrentTurn + 1) % len(gs.Players)
	}
}

func (gs *GameState) endGame(winnerIdx int) {
	gs.GameOver = true
	gs.Winner = winnerIdx

	// Score: each loser gets deadwood points added to their score
	for i := range gs.Players {
		if i == winnerIdx {
			continue
		}
		for _, c := range gs.Players[i].Hand {
			gs.Players[i].Score += c.Points()
		}
	}
}

// reshuffleDiscard takes all cards from the discard pile except the top,
// shuffles them, and puts them back as the deck.
func (gs *GameState) reshuffleDiscard() {
	if len(gs.DiscardPile) <= 1 {
		return
	}
	top := gs.DiscardPile[len(gs.DiscardPile)-1]
	gs.Deck = &Deck{cards: gs.DiscardPile[:len(gs.DiscardPile)-1]}
	gs.Deck.Shuffle()
	gs.DiscardPile = []Card{top}
}
