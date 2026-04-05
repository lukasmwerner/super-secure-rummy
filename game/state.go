package game

import (
	"sync"

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

	MeldLen []int

	Players []string
	Updates []chan struct{}
	Mu      sync.Mutex
}

func (s *State) Notify() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	for _, c := range s.Updates {
		select {
		case c <- struct{}{}:
		default:
		}
	}
}

type GameManager struct {
	states map[uint64]*State
	mu     sync.Mutex

	pendingGame *State
	pendingID   uint64
}

func NewGameManager() *GameManager {
	return &GameManager{
		states: make(map[uint64]*State),
	}
}

func (gm *GameManager) GetOrCreateGame(playerPubKey string) (*State, uint64, chan struct{}) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	updateChan := make(chan struct{}, 1)

	if gm.pendingGame != nil {
		s := gm.pendingGame
		id := gm.pendingID

		s.Mu.Lock()
		s.Players = append(s.Players, playerPubKey)
		s.Updates = append(s.Updates, updateChan)
		s.Mu.Unlock()

		gm.pendingGame = nil
		gm.pendingID = 0
		s.Notify()
		return s, id, updateChan
	}

	// Create new game
	id := uint64(len(gm.states) + 1)
	s := &State{
		Hand:    map[string][]*lipgloss.Layer{},
		Discard: []*lipgloss.Layer{},
		Draw:    []*lipgloss.Layer{},
		Melds:   map[string][]*lipgloss.Layer{},
		MeldLen: make([]int, 12),
		Players: []string{playerPubKey},
		Updates: []chan struct{}{updateChan},
	}
	gm.states[id] = s
	gm.pendingGame = s
	gm.pendingID = id

	return s, id, updateChan
}
