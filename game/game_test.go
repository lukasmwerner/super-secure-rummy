package game

import (
	"testing"
)

func TestNewDeck(t *testing.T) {
	d := NewDeck()
	if d.Remaining() != 52 {
		t.Fatalf("expected 52 cards, got %d", d.Remaining())
	}

	// Verify all 52 unique cards exist
	seen := make(map[Card]bool)
	for i := 0; i < 52; i++ {
		card, ok := d.Draw()
		if !ok {
			t.Fatalf("failed to draw card %d", i)
		}
		if seen[card] {
			t.Fatalf("duplicate card: %s", card)
		}
		seen[card] = true
	}

	if !d.IsEmpty() {
		t.Fatal("deck should be empty after drawing 52 cards")
	}

	if _, ok := d.Draw(); ok {
		t.Fatal("should not be able to draw from empty deck")
	}
}

func TestCardPoints(t *testing.T) {
	tests := []struct {
		card   Card
		points int
	}{
		{Card{Spades, 1}, 1},     // Ace
		{Card{Hearts, 5}, 5},     // 5
		{Card{Diamonds, 9}, 9},   // 9
		{Card{Clubs, 10}, 10},    // 10
		{Card{Spades, 11}, 10},   // Jack
		{Card{Hearts, 12}, 10},   // Queen
		{Card{Diamonds, 13}, 10}, // King
	}
	for _, tt := range tests {
		if got := tt.card.Points(); got != tt.points {
			t.Errorf("%s.Points() = %d, want %d", tt.card, got, tt.points)
		}
	}
}

func TestValidateSet(t *testing.T) {
	// Valid 3-card set
	if !ValidateSet([]Card{{Spades, 5}, {Hearts, 5}, {Diamonds, 5}}) {
		t.Error("3 fives should be a valid set")
	}

	// Valid 4-card set
	if !ValidateSet([]Card{{Spades, 10}, {Hearts, 10}, {Diamonds, 10}, {Clubs, 10}}) {
		t.Error("4 tens should be a valid set")
	}

	// Invalid: different ranks
	if ValidateSet([]Card{{Spades, 5}, {Hearts, 6}, {Diamonds, 5}}) {
		t.Error("mixed ranks should not be a valid set")
	}

	// Invalid: duplicate suit
	if ValidateSet([]Card{{Spades, 5}, {Spades, 5}, {Hearts, 5}}) {
		t.Error("duplicate suits should not be a valid set")
	}

	// Invalid: too few
	if ValidateSet([]Card{{Spades, 5}, {Hearts, 5}}) {
		t.Error("2 cards should not be a valid set")
	}

	// Invalid: too many
	if ValidateSet([]Card{{Spades, 5}, {Hearts, 5}, {Diamonds, 5}, {Clubs, 5}, {Spades, 5}}) {
		t.Error("5 cards should not be a valid set")
	}
}

func TestValidateRun(t *testing.T) {
	// Valid 3-card run
	if !ValidateRun([]Card{{Hearts, 3}, {Hearts, 4}, {Hearts, 5}}) {
		t.Error("3-4-5 of hearts should be a valid run")
	}

	// Valid run with Ace low
	if !ValidateRun([]Card{{Spades, 1}, {Spades, 2}, {Spades, 3}}) {
		t.Error("A-2-3 of spades should be a valid run")
	}

	// Valid run, unordered input
	if !ValidateRun([]Card{{Diamonds, 7}, {Diamonds, 5}, {Diamonds, 6}}) {
		t.Error("unordered 5-6-7 should be valid run")
	}

	// Valid long run
	if !ValidateRun([]Card{{Clubs, 8}, {Clubs, 9}, {Clubs, 10}, {Clubs, 11}, {Clubs, 12}}) {
		t.Error("8-9-10-J-Q should be valid run")
	}

	// Invalid: Q-K-A (ace is low only)
	if ValidateRun([]Card{{Hearts, 12}, {Hearts, 13}, {Hearts, 1}}) {
		t.Error("Q-K-A should not be valid (ace is low only)")
	}

	// Invalid: gap
	if ValidateRun([]Card{{Hearts, 3}, {Hearts, 5}, {Hearts, 6}}) {
		t.Error("3-5-6 should not be a valid run (gap)")
	}

	// Invalid: mixed suits
	if ValidateRun([]Card{{Hearts, 3}, {Spades, 4}, {Hearts, 5}}) {
		t.Error("mixed suits should not be a valid run")
	}

	// Invalid: too few
	if ValidateRun([]Card{{Hearts, 3}, {Hearts, 4}}) {
		t.Error("2 cards should not be a valid run")
	}
}

func TestCanLayOff(t *testing.T) {
	// Lay off on a set
	set := Meld{Type: MeldSet, Cards: []Card{{Spades, 7}, {Hearts, 7}, {Diamonds, 7}}}
	if !CanLayOff(Card{Clubs, 7}, set) {
		t.Error("clubs 7 should lay off on set of 7s")
	}
	if CanLayOff(Card{Spades, 7}, set) {
		t.Error("spades 7 already in set, should not lay off")
	}
	if CanLayOff(Card{Clubs, 8}, set) {
		t.Error("wrong rank should not lay off on set")
	}

	// Full set (4 cards) — no more lay-offs
	fullSet := Meld{Type: MeldSet, Cards: []Card{{Spades, 7}, {Hearts, 7}, {Diamonds, 7}, {Clubs, 7}}}
	if CanLayOff(Card{Spades, 7}, fullSet) {
		t.Error("should not lay off on a full 4-card set")
	}

	// Lay off on a run
	run := Meld{Type: MeldRun, Cards: []Card{{Hearts, 5}, {Hearts, 6}, {Hearts, 7}}}
	if !CanLayOff(Card{Hearts, 4}, run) {
		t.Error("4 of hearts should extend run at low end")
	}
	if !CanLayOff(Card{Hearts, 8}, run) {
		t.Error("8 of hearts should extend run at high end")
	}
	if CanLayOff(Card{Hearts, 3}, run) {
		t.Error("3 of hearts should not extend 5-6-7 run")
	}
	if CanLayOff(Card{Spades, 4}, run) {
		t.Error("wrong suit should not extend run")
	}
}

func TestNewGameState(t *testing.T) {
	gs := NewGameState([]string{"alice", "bob"})

	if !gs.Started {
		t.Error("game should be started")
	}
	if gs.GameOver {
		t.Error("game should not be over")
	}
	if len(gs.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(gs.Players))
	}
	for _, p := range gs.Players {
		if len(p.Hand) != 7 {
			t.Errorf("player %s should have 7 cards, got %d", p.ID, len(p.Hand))
		}
	}

	// 52 - 14 dealt - 1 discard = 37 remaining
	if gs.Deck.Remaining() != 37 {
		t.Errorf("deck should have 37 remaining, got %d", gs.Deck.Remaining())
	}
	if len(gs.DiscardPile) != 1 {
		t.Errorf("discard pile should have 1 card, got %d", len(gs.DiscardPile))
	}
	if gs.Phase != PhaseDraw {
		t.Error("should start in draw phase")
	}
}

func TestApplyDrawAndDiscard(t *testing.T) {
	gs := NewGameState([]string{"alice", "bob"})
	alice := gs.Players[0].ID
	bob := gs.Players[1].ID

	// Bob can't draw — not his turn
	err := gs.Apply(GameAction{PlayerID: bob, Type: ActionDrawFromDeck})
	if err != ErrNotYourTurn {
		t.Errorf("expected ErrNotYourTurn, got %v", err)
	}

	// Alice draws from deck
	err = gs.Apply(GameAction{PlayerID: alice, Type: ActionDrawFromDeck})
	if err != nil {
		t.Fatalf("alice draw: %v", err)
	}
	if len(gs.Players[0].Hand) != 8 {
		t.Errorf("alice should have 8 cards after draw, got %d", len(gs.Players[0].Hand))
	}
	if gs.Phase != PhaseAction {
		t.Error("should be in action phase after draw")
	}

	// Alice can't draw again
	err = gs.Apply(GameAction{PlayerID: alice, Type: ActionDrawFromDeck})
	if err != ErrWrongPhase {
		t.Errorf("expected ErrWrongPhase for double draw, got %v", err)
	}

	// Alice discards
	discardCard := gs.Players[0].Hand[0]
	err = gs.Apply(GameAction{PlayerID: alice, Type: ActionDiscard, Cards: []Card{discardCard}})
	if err != nil {
		t.Fatalf("alice discard: %v", err)
	}
	if len(gs.Players[0].Hand) != 7 {
		t.Errorf("alice should have 7 cards after discard, got %d", len(gs.Players[0].Hand))
	}

	// Now it's bob's turn
	if gs.CurrentTurn != 1 {
		t.Error("should be bob's turn now")
	}
	if gs.Phase != PhaseDraw {
		t.Error("bob should be in draw phase")
	}
}

func TestApplyDrawFromDiscard(t *testing.T) {
	gs := NewGameState([]string{"alice", "bob"})
	alice := gs.Players[0].ID

	topDiscard := gs.DiscardPile[len(gs.DiscardPile)-1]

	err := gs.Apply(GameAction{PlayerID: alice, Type: ActionDrawFromDiscard})
	if err != nil {
		t.Fatalf("draw from discard: %v", err)
	}

	// Alice should now have the top discard card
	found := false
	for _, c := range gs.Players[0].Hand {
		if c.Equal(topDiscard) {
			found = true
			break
		}
	}
	if !found {
		t.Error("alice should have the card drawn from discard")
	}
}

func TestApplyMeld(t *testing.T) {
	gs := NewGameState([]string{"alice", "bob"})

	// Give Alice a known set: three 5s
	gs.Players[0].Hand = []Card{
		{Spades, 5}, {Hearts, 5}, {Diamonds, 5},
		{Clubs, 1}, {Clubs, 2}, {Clubs, 3}, {Clubs, 4},
	}

	// Draw first
	alice := gs.Players[0].ID
	_ = gs.Apply(GameAction{PlayerID: alice, Type: ActionDrawFromDeck})

	// Meld the three 5s
	err := gs.Apply(GameAction{
		PlayerID: alice,
		Type:     ActionMeld,
		Cards:    []Card{{Spades, 5}, {Hearts, 5}, {Diamonds, 5}},
	})
	if err != nil {
		t.Fatalf("meld: %v", err)
	}
	if len(gs.Melds) != 1 {
		t.Errorf("expected 1 meld on table, got %d", len(gs.Melds))
	}
	if len(gs.Players[0].Hand) != 5 { // 7 + 1 draw - 3 melded = 5
		t.Errorf("alice should have 5 cards, got %d", len(gs.Players[0].Hand))
	}
}

func TestSnapshotHidesOpponentHands(t *testing.T) {
	gs := NewGameState([]string{"alice", "bob"})

	aliceSnap := gs.SnapshotFor("alice")
	if len(aliceSnap.YourHand) != 7 {
		t.Errorf("alice should see 7 cards in her hand, got %d", len(aliceSnap.YourHand))
	}

	// Alice should not see bob's hand
	for _, p := range aliceSnap.Players {
		if p.ID == "bob" {
			if p.CardCount != 7 {
				t.Errorf("alice should see bob has 7 cards, got %d", p.CardCount)
			}
		}
	}

	// Bob's snapshot should show his own hand
	bobSnap := gs.SnapshotFor("bob")
	if len(bobSnap.YourHand) != 7 {
		t.Errorf("bob should see 7 cards in his hand, got %d", len(bobSnap.YourHand))
	}

	// Verify snapshots show different hands
	aliceHand := aliceSnap.YourHand
	bobHand := bobSnap.YourHand
	allSame := true
	for i := range aliceHand {
		if i >= len(bobHand) || !aliceHand[i].Equal(bobHand[i]) {
			allSame = false
			break
		}
	}
	if allSame && len(aliceHand) == len(bobHand) {
		t.Error("alice and bob should have different hands (extremely unlikely to be identical)")
	}
}

func TestWinCondition(t *testing.T) {
	gs := NewGameState([]string{"alice", "bob"})

	// Give Alice a minimal hand that she can empty
	gs.Players[0].Hand = []Card{
		{Spades, 5}, {Hearts, 5}, {Diamonds, 5}, {Clubs, 9},
	}

	alice := gs.Players[0].ID

	// Draw
	_ = gs.Apply(GameAction{PlayerID: alice, Type: ActionDrawFromDeck})

	// Meld the three 5s
	_ = gs.Apply(GameAction{
		PlayerID: alice,
		Type:     ActionMeld,
		Cards:    []Card{{Spades, 5}, {Hearts, 5}, {Diamonds, 5}},
	})

	// Discard the remaining cards (she has 9♣ + whatever she drew)
	// Discard the drawn card first
	if len(gs.Players[0].Hand) == 2 {
		// Discard drawn card
		err := gs.Apply(GameAction{
			PlayerID: alice,
			Type:     ActionDiscard,
			Cards:    []Card{gs.Players[0].Hand[1]},
		})
		if err != nil {
			t.Fatalf("discard drawn card: %v", err)
		}
	}

	// If alice still has a card, it's now bob's turn, so this test checks
	// the win condition in the meld case. Let's set up a cleaner scenario.
}

func TestWinByEmptyHand(t *testing.T) {
	gs := NewGameState([]string{"alice", "bob"})

	alice := gs.Players[0].ID

	// Set Alice's hand to exactly 3 cards that form a valid set + 1 to discard
	gs.Players[0].Hand = []Card{
		{Spades, 7}, {Hearts, 7}, {Diamonds, 7}, {Clubs, 2},
	}

	// Draw
	_ = gs.Apply(GameAction{PlayerID: alice, Type: ActionDrawFromDeck})

	// Meld the set
	err := gs.Apply(GameAction{
		PlayerID: alice,
		Type:     ActionMeld,
		Cards:    []Card{{Spades, 7}, {Hearts, 7}, {Diamonds, 7}},
	})
	if err != nil {
		t.Fatalf("meld: %v", err)
	}

	// Hand should now have 2♣ + drawn card = 2 cards
	// If melding emptied her hand (it won't here since she has 2♣ + drawn card)
	// Discard one
	discardCard := gs.Players[0].Hand[0]
	err = gs.Apply(GameAction{
		PlayerID: alice,
		Type:     ActionDiscard,
		Cards:    []Card{discardCard},
	})
	if err != nil {
		t.Fatalf("discard: %v", err)
	}

	// If she has 1 card left after discard, it's not a win yet
	// Let's test the clean win case: meld empties the hand
	gs2 := NewGameState([]string{"alice", "bob"})
	gs2.Players[0].Hand = []Card{
		{Spades, 7}, {Hearts, 7}, {Diamonds, 7},
	}
	gs2.Phase = PhaseAction // Skip draw for testing

	err = gs2.Apply(GameAction{
		PlayerID: "alice",
		Type:     ActionMeld,
		Cards:    []Card{{Spades, 7}, {Hearts, 7}, {Diamonds, 7}},
	})
	if err != nil {
		t.Fatalf("meld for win: %v", err)
	}
	if !gs2.GameOver {
		t.Error("game should be over when hand is empty after meld")
	}
	if gs2.Players[gs2.Winner].ID != "alice" {
		t.Error("alice should be the winner")
	}
}
