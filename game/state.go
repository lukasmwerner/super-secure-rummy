package game

import (
	"charm.land/lipgloss/v2"
)

type Game struct {
	State string
	ID    uint64
}

type State struct {
	Hand    map[string][]*lipgloss.Layer
	Discard []*lipgloss.Layer
	Draw    []*lipgloss.Layer
	Melds   map[string][]*lipgloss.Layer
	Turn    string
}
