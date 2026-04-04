package game

import (
	"github.com/lukasmwerner/secure-rummy/card"
)

type Game struct {
	State string
	ID    uint64
}

type State struct {
	Hand    map[string][]card.Card
	Discard []card.Card
	Draw    []card.Card
	Melds   map[string][]card.Card
}
