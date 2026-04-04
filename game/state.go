package game

import (
	"github.com/charmbracelet/ssh"
	"github.com/lukasmwerner/secure-rummy/card"
)

type Game struct {
	State string
	ID    uint64
}

type State struct {
	Hand    map[ssh.PublicKey][]card.Card
	Discard []card.Card
	Draw    []card.Card
	Melds   map[ssh.PublicKey][]card.Card
}
