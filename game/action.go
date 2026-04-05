package game

// ActionType enumerates every legal player action.
type ActionType int

const (
	ActionDrawFromDeck    ActionType = iota // Draw one card from the stock pile
	ActionDrawFromDiscard                   // Draw the top card from the discard pile
	ActionDiscard                           // Discard one card to end turn
	ActionMeld                              // Lay down a new meld (set or run)
	ActionLayOff                            // Add card(s) to an existing meld
)

// GameAction represents a player's intent, sent to the Hub for processing.
type GameAction struct {
	PlayerID   string // SSH public key fingerprint
	Type       ActionType
	Cards      []Card // Cards involved (for discard, meld, lay-off)
	TargetMeld int    // Index of target meld (for lay-off only)
}
