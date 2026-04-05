package game

import (
	"fmt"
)

// PlayerConn links a player's identity to their BubbleTea program's Send function.
// The Send function injects tea.Msg values into the player's event loop.
type PlayerConn struct {
	ID   string
	Name string
	Send func(msg interface{}) // Calls tea.Program.Send() — thread-safe
}

// Hub is the central game coordinator running in its own goroutine.
// All game state mutations flow through the Hub's action channel,
// ensuring single-goroutine ownership of GameState (actor pattern).
type Hub struct {
	Actions    chan GameAction
	Register   chan *PlayerConn
	Unregister chan string // PlayerID to unregister

	players map[string]*PlayerConn
	state   *GameState

	minPlayers int
	maxPlayers int
}

// NewHub creates a new Hub ready to accept player connections.
func NewHub(minPlayers, maxPlayers int) *Hub {
	return &Hub{
		Actions:    make(chan GameAction, 64),
		Register:   make(chan *PlayerConn, 8),
		Unregister: make(chan string, 8),
		players:    make(map[string]*PlayerConn),
		minPlayers: minPlayers,
		maxPlayers: maxPlayers,
	}
}

// Run is the main event loop. Must be called as `go hub.Run()`.
// It owns all game state — no other goroutine reads or writes GameState.
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.handleRegister(conn)

		case playerID := <-h.Unregister:
			h.handleUnregister(playerID)

		case action := <-h.Actions:
			h.handleAction(action)
		}
	}
}

func (h *Hub) handleRegister(conn *PlayerConn) {
	// If the game is over, reset state so a new lobby can form
	if h.state != nil && h.state.GameOver {
		h.state = nil
	}

	// Check if this is a reconnecting player to an active game
	if h.state != nil && h.state.Started {
		for i := range h.state.Players {
			if h.state.Players[i].ID == conn.ID {
				// Reconnecting — restore connection
				h.state.Players[i].Connected = true
				h.players[conn.ID] = conn
				h.broadcastSnapshots()
				return
			}
		}
		// Game in progress and this player isn't part of it
		conn.Send(GameErrorMsg{Error: "game already in progress"})
		return
	}

	// Lobby phase — accept new players
	if len(h.players) >= h.maxPlayers {
		conn.Send(GameErrorMsg{Error: "game is full"})
		return
	}

	h.players[conn.ID] = conn
	h.broadcastLobby()
}

func (h *Hub) handleUnregister(playerID string) {
	if _, ok := h.players[playerID]; !ok {
		return
	}

	delete(h.players, playerID)

	if h.state != nil {
		for i := range h.state.Players {
			if h.state.Players[i].ID == playerID {
				h.state.Players[i].Connected = false
				break
			}
		}

		// If all players disconnected, reset everything for next session
		if len(h.players) == 0 {
			h.state = nil
		} else {
			h.broadcastSnapshots()
		}
	} else {
		// Still in lobby
		if len(h.players) == 0 {
			h.state = nil
		} else {
			h.broadcastLobby()
		}
	}
}

func (h *Hub) handleAction(action GameAction) {
	// Special action: start game
	if action.Type == ActionStartGame {
		h.handleStartGame(action.PlayerID)
		return
	}

	// Special action: new game (from game over screen)
	if action.Type == ActionNewGame {
		h.handleNewGame(action.PlayerID)
		return
	}

	if h.state == nil || !h.state.Started {
		h.sendError(action.PlayerID, "game has not started")
		return
	}

	if err := h.state.Apply(action); err != nil {
		h.sendError(action.PlayerID, err.Error())
		return
	}

	h.broadcastSnapshots()
}

func (h *Hub) handleStartGame(requestingPlayer string) {
	if h.state != nil && h.state.Started && !h.state.GameOver {
		h.sendError(requestingPlayer, "game already started")
		return
	}

	// If game is over, reset first
	if h.state != nil && h.state.GameOver {
		h.state = nil
	}

	if len(h.players) < h.minPlayers {
		h.sendError(requestingPlayer, fmt.Sprintf("need at least %d players to start", h.minPlayers))
		return
	}

	// Collect player IDs
	ids := make([]string, 0, len(h.players))
	for id := range h.players {
		ids = append(ids, id)
	}

	h.state = NewGameState(ids)
	h.broadcastSnapshots()
}

// handleNewGame resets the game and returns all connected players to the lobby.
func (h *Hub) handleNewGame(requestingPlayer string) {
	h.state = nil
	h.broadcastLobby()
}

// broadcastSnapshots sends a personalized GameSnapshot to every connected player.
func (h *Hub) broadcastSnapshots() {
	if h.state == nil {
		return
	}
	for id, conn := range h.players {
		snapshot := h.state.SnapshotFor(id)
		conn.Send(GameSnapshotMsg(snapshot))
	}
}

// broadcastLobby sends lobby status to all connected players.
func (h *Hub) broadcastLobby() {
	players := make([]PlayerInfo, 0, len(h.players))
	for _, conn := range h.players {
		players = append(players, PlayerInfo{
			ID:        conn.ID,
			Name:      conn.Name,
			Connected: true,
		})
	}
	for _, conn := range h.players {
		conn.Send(LobbyMsg{
			Players:    players,
			MinPlayers: h.minPlayers,
			MaxPlayers: h.maxPlayers,
		})
	}
}

func (h *Hub) sendError(playerID string, msg string) {
	if conn, ok := h.players[playerID]; ok {
		conn.Send(GameErrorMsg{Error: msg})
	}
}

// --- Additional action type for starting a game ---

const ActionStartGame ActionType = 100
const ActionNewGame ActionType = 101

// --- Message types sent from Hub to clients ---

// GameSnapshotMsg wraps GameSnapshot for BubbleTea's type-switch.
type GameSnapshotMsg GameSnapshot

// GameErrorMsg is sent to a player when their action is illegal.
type GameErrorMsg struct {
	Error string
}

// LobbyMsg is sent to players while waiting for the game to start.
type LobbyMsg struct {
	Players    []PlayerInfo
	MinPlayers int
	MaxPlayers int
}
