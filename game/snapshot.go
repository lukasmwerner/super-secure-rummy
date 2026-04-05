package game

// GameSnapshot is a read-only, per-player view of the game state.
// This is the message sent to each player's BubbleTea program after every state change.
type GameSnapshot struct {
	YourHand      []Card       // Only this player's hand
	DiscardTop    *Card        // Top card of discard pile (nil if empty)
	DiscardCount  int          // Total cards in discard pile
	DeckRemaining int          // Cards left in the stock pile
	Melds         []Meld       // All melds on the table
	Players       []PlayerInfo // All players' public info (in turn order)
	CurrentTurn   string       // PlayerID of whose turn it is
	YourTurn      bool         // Convenience flag
	Phase         TurnPhase    // Current turn phase (draw/action/discard)
	GameOver      bool
	Winner        string // PlayerID of the winner (empty if not over)
	Scores        map[string]int
	Started       bool   // Whether the game has started
	StatusMessage string // Optional status/error for this player
}

// PlayerInfo is the public information about a player visible to all.
type PlayerInfo struct {
	ID        string
	Name      string
	CardCount int
	Score     int
	Connected bool
	IsYou     bool
}
