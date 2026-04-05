package client

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lukasmwerner/super-secure-rummy/card"
	"github.com/lukasmwerner/super-secure-rummy/game"
)

var (
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	turnStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Bold(true)
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))
	selectedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF79C6"))
	normalBorder  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
)

// Screen represents the current UI screen.
type Screen int

const (
	ScreenLobby Screen = iota
	ScreenGame
	ScreenGameOver
)

// FocusArea represents which area of the game UI is focused.
type FocusArea int

const (
	FocusHand FocusArea = iota
	FocusDraw
	FocusDiscard
	FocusMelds
)

var focusNames = [...]string{"Hand", "Draw", "Discard", "Melds"}

// Model is the BubbleTea model for each connected SSH client.
type Model struct {
	PlayerID string
	Width    int
	Height   int
	Bg       string
	Hub      *game.Hub

	// UI state
	Screen     Screen
	Focus      FocusArea
	CursorPos  int    // Selected card index within hand
	MeldCursor int    // Selected meld index
	StatusMsg  string // Temporary status/error message

	// Game state (from snapshots)
	Snapshot game.GameSnapshot

	// Lobby state
	LobbyPlayers []game.PlayerInfo
	MinPlayers   int
	MaxPlayers   int

	// Card selection for melding
	Selected   map[int]bool // Indices of selected cards in hand
	SelectMode bool         // Whether we're in multi-select mode for melding

	// Connection tracking
	Conn *game.PlayerConn
}

func (m Model) Init() tea.Cmd {
	// Registration with the Hub happens in main.go's makeProgramHandler,
	// before the program starts. Here we just detect terminal capabilities.
	return tea.RequestBackgroundColor
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case game.GameSnapshotMsg:
		m.Snapshot = game.GameSnapshot(msg)
		if m.Snapshot.GameOver {
			m.Screen = ScreenGameOver
		} else {
			m.Screen = ScreenGame
		}
		m.StatusMsg = m.Snapshot.StatusMessage
		// Clamp cursor positions
		if m.CursorPos >= len(m.Snapshot.YourHand) {
			m.CursorPos = max(0, len(m.Snapshot.YourHand)-1)
		}
		if m.MeldCursor >= len(m.Snapshot.Melds) {
			m.MeldCursor = max(0, len(m.Snapshot.Melds)-1)
		}
		return m, nil

	case game.GameErrorMsg:
		m.StatusMsg = "Error: " + msg.Error
		return m, nil

	case game.LobbyMsg:
		m.Screen = ScreenLobby
		m.LobbyPlayers = msg.Players
		m.MinPlayers = msg.MinPlayers
		m.MaxPlayers = msg.MaxPlayers
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.BackgroundColorMsg:
		if msg.IsDark() {
			m.Bg = "dark"
		} else {
			m.Bg = "light"
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "q", "ctrl+c":
		// Unregister from hub before quitting
		m.Hub.Unregister <- m.PlayerID
		return m, tea.Quit
	}

	switch m.Screen {
	case ScreenLobby:
		return m.handleLobbyKey(key)
	case ScreenGame:
		return m.handleGameKey(key)
	case ScreenGameOver:
		return m.handleGameOverKey(key)
	}
	return m, nil
}

func (m Model) handleLobbyKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter", "s":
		// Start game
		return m, func() tea.Msg {
			m.Hub.Actions <- game.GameAction{
				PlayerID: m.PlayerID,
				Type:     game.ActionStartGame,
			}
			return nil
		}
	}
	return m, nil
}

func (m Model) handleGameKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "tab":
		m.Focus = (m.Focus + 1) % 4
		m.SelectMode = false
		m.Selected = nil
		return m, nil

	case "shift+tab":
		m.Focus = (m.Focus + 3) % 4 // Wrap backwards
		m.SelectMode = false
		m.Selected = nil
		return m, nil

	case "left", "h":
		switch m.Focus {
		case FocusHand:
			if m.CursorPos > 0 {
				m.CursorPos--
			}
		case FocusMelds:
			if m.MeldCursor > 0 {
				m.MeldCursor--
			}
		}
		return m, nil

	case "right", "l":
		switch m.Focus {
		case FocusHand:
			if m.CursorPos < len(m.Snapshot.YourHand)-1 {
				m.CursorPos++
			}
		case FocusMelds:
			if m.MeldCursor < len(m.Snapshot.Melds)-1 {
				m.MeldCursor++
			}
		}
		return m, nil

	case "space": // Space — toggle card selection for melding
		if m.Focus == FocusHand && len(m.Snapshot.YourHand) > 0 {
			if m.Selected == nil {
				m.Selected = make(map[int]bool)
			}
			m.SelectMode = true
			m.Selected[m.CursorPos] = !m.Selected[m.CursorPos]
			if !m.Selected[m.CursorPos] {
				delete(m.Selected, m.CursorPos)
			}
			if len(m.Selected) == 0 {
				m.SelectMode = false
			}
		}
		return m, nil

	case "enter":
		return m.handleAction()

	case "d": // Draw from deck
		if !m.Snapshot.YourTurn || m.Snapshot.Phase != game.PhaseDraw {
			m.StatusMsg = "Can't draw right now"
			return m, nil
		}
		return m, func() tea.Msg {
			m.Hub.Actions <- game.GameAction{
				PlayerID: m.PlayerID,
				Type:     game.ActionDrawFromDeck,
			}
			return nil
		}

	case "x": // Discard the cursor card
		if !m.Snapshot.YourTurn {
			m.StatusMsg = "Not your turn"
			return m, nil
		}
		if m.Snapshot.Phase != game.PhaseAction && m.Snapshot.Phase != game.PhaseDiscard {
			m.StatusMsg = "Draw a card first"
			return m, nil
		}
		if m.Focus == FocusHand && len(m.Snapshot.YourHand) > 0 {
			selectedCard := m.Snapshot.YourHand[m.CursorPos]
			return m, func() tea.Msg {
				m.Hub.Actions <- game.GameAction{
					PlayerID: m.PlayerID,
					Type:     game.ActionDiscard,
					Cards:    []game.Card{selectedCard},
				}
				return nil
			}
		}

	case "escape":
		m.SelectMode = false
		m.Selected = nil
		m.StatusMsg = ""
		return m, nil
	}

	return m, nil
}

func (m Model) handleAction() (tea.Model, tea.Cmd) {
	if !m.Snapshot.YourTurn {
		m.StatusMsg = "Not your turn"
		return m, nil
	}

	switch m.Focus {
	case FocusHand:
		if m.SelectMode && len(m.Selected) >= 3 {
			// Meld selected cards
			cards := make([]game.Card, 0, len(m.Selected))
			for idx := range m.Selected {
				if idx < len(m.Snapshot.YourHand) {
					cards = append(cards, m.Snapshot.YourHand[idx])
				}
			}
			m.SelectMode = false
			m.Selected = nil
			return m, func() tea.Msg {
				m.Hub.Actions <- game.GameAction{
					PlayerID: m.PlayerID,
					Type:     game.ActionMeld,
					Cards:    cards,
				}
				return nil
			}
		} else if m.SelectMode && len(m.Selected) > 0 {
			// Not enough cards selected for a meld
			m.StatusMsg = fmt.Sprintf("Need at least 3 cards for a meld (have %d selected)", len(m.Selected))
		} else if !m.SelectMode && len(m.Snapshot.YourHand) > 0 {
			// Enter without selection = discard cursor card
			if m.Snapshot.Phase == game.PhaseAction || m.Snapshot.Phase == game.PhaseDiscard {
				selectedCard := m.Snapshot.YourHand[m.CursorPos]
				return m, func() tea.Msg {
					m.Hub.Actions <- game.GameAction{
						PlayerID: m.PlayerID,
						Type:     game.ActionDiscard,
						Cards:    []game.Card{selectedCard},
					}
					return nil
				}
			} else {
				m.StatusMsg = "Draw a card first before discarding"
			}
		}

	case FocusDraw:
		if m.Snapshot.Phase == game.PhaseDraw {
			return m, func() tea.Msg {
				m.Hub.Actions <- game.GameAction{
					PlayerID: m.PlayerID,
					Type:     game.ActionDrawFromDeck,
				}
				return nil
			}
		}

	case FocusDiscard:
		if m.Snapshot.Phase == game.PhaseDraw {
			return m, func() tea.Msg {
				m.Hub.Actions <- game.GameAction{
					PlayerID: m.PlayerID,
					Type:     game.ActionDrawFromDiscard,
				}
				return nil
			}
		}

	case FocusMelds:
		// Lay off the cursor card onto the selected meld
		if m.Snapshot.Phase == game.PhaseAction && len(m.Snapshot.YourHand) > 0 {
			selectedCard := m.Snapshot.YourHand[m.CursorPos]
			return m, func() tea.Msg {
				m.Hub.Actions <- game.GameAction{
					PlayerID:   m.PlayerID,
					Type:       game.ActionLayOff,
					Cards:      []game.Card{selectedCard},
					TargetMeld: m.MeldCursor,
				}
				return nil
			}
		}
	}

	return m, nil
}

func (m Model) handleGameOverKey(key string) (tea.Model, tea.Cmd) {
	return m, nil
}

// --- VIEW ---

func (m Model) View() tea.View {
	switch m.Screen {
	case ScreenLobby:
		return m.viewLobby()
	case ScreenGame:
		return m.viewGame()
	case ScreenGameOver:
		return m.viewGameOver()
	}

	v := tea.NewView("Loading...")
	v.AltScreen = true
	return v
}

func (m Model) viewLobby() tea.View {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("♠ ♥ Secure Rummy ♦ ♣") + "\n\n")
	sb.WriteString(infoStyle.Render("Waiting for players...") + "\n\n")

	for i, p := range m.LobbyPlayers {
		marker := "  "
		if p.ID == m.PlayerID {
			marker = "→ "
		}
		sb.WriteString(fmt.Sprintf("%s%d. %s", marker, i+1, p.Name))
		if p.ID == m.PlayerID {
			sb.WriteString(" (you)")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n%d/%d players (need %d)\n", len(m.LobbyPlayers), m.MaxPlayers, m.MinPlayers))

	if len(m.LobbyPlayers) >= m.MinPlayers {
		sb.WriteString(turnStyle.Render("\nPress [s] or [enter] to start the game!") + "\n")
	}
	sb.WriteString(helpStyle.Render("\nPress [q] to quit"))

	content := lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, sb.String())
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) viewGame() tea.View {
	var sb strings.Builder

	// Header: turn indicator + phase
	if m.Snapshot.YourTurn {
		sb.WriteString(turnStyle.Render("★ YOUR TURN") + " ")
		switch m.Snapshot.Phase {
		case game.PhaseDraw:
			sb.WriteString(infoStyle.Render("(Draw a card: [d]=deck, Tab to Discard pile + [enter])"))
		case game.PhaseAction:
			sb.WriteString(infoStyle.Render("(Meld/LayOff or Discard: [space]=select, [enter]=confirm)"))
		case game.PhaseDiscard:
			sb.WriteString(infoStyle.Render("(Discard a card to end turn)"))
		}
	} else {
		sb.WriteString(infoStyle.Render("Waiting for " + m.Snapshot.CurrentTurn[:min(8, len(m.Snapshot.CurrentTurn))] + "'s turn..."))
	}
	sb.WriteString("\n\n")

	// Opponent info
	for _, p := range m.Snapshot.Players {
		if p.IsYou {
			continue
		}
		connStatus := "●"
		if !p.Connected {
			connStatus = "○"
		}
		isTurn := ""
		if p.ID == m.Snapshot.CurrentTurn {
			isTurn = " ←"
		}
		sb.WriteString(fmt.Sprintf("  %s %s: %d cards%s\n", connStatus, p.Name[:min(8, len(p.Name))], p.CardCount, isTurn))
	}
	sb.WriteString("\n")

	// Table: Draw pile + Discard pile
	drawPile := card.DrawPile()
	var discardPile *lipgloss.Layer
	if m.Snapshot.DiscardTop != nil {
		c := *m.Snapshot.DiscardTop
		suit := convertSuit(c.Suit)
		discardPile = card.CardLayer(suit, strconv.Itoa(c.Rank), m.Bg, card.HighlightNone)
	} else {
		discardPile = card.PlaceHolderPile()
	}

	drawLabel := fmt.Sprintf(" Deck (%d)", m.Snapshot.DeckRemaining)
	discardLabel := fmt.Sprintf(" Discard (%d)", m.Snapshot.DiscardCount)

	drawBox := lipgloss.NewCompositor(drawPile).Render()
	discardBox := lipgloss.NewCompositor(discardPile).Render()

	if m.Focus == FocusDraw {
		drawBox = selectedStyle.Render(drawBox)
	} else {
		drawBox = normalBorder.Render(drawBox)
	}
	if m.Focus == FocusDiscard {
		discardBox = selectedStyle.Render(discardBox)
	} else {
		discardBox = normalBorder.Render(discardBox)
	}

	piles := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Center, drawLabel, drawBox),
		"  ",
		lipgloss.JoinVertical(lipgloss.Center, discardLabel, discardBox),
	)
	sb.WriteString(piles + "\n\n")

	// Melds on the table
	if len(m.Snapshot.Melds) > 0 {
		sb.WriteString(infoStyle.Render("Melds:") + "\n")

		var meldRenders []string
		for i, meld := range m.Snapshot.Melds {
			var meldLayers []*lipgloss.Layer
			for j, c := range meld.Cards {
				suit := convertSuit(c.Suit)
				if j == len(meld.Cards)-1 {
					meldLayers = append(meldLayers, card.CardLayer(suit, strconv.Itoa(c.Rank), m.Bg, card.HighlightNone))
				} else {
					meldLayers = append(meldLayers, card.HalfCardLayer(false, suit, strconv.Itoa(c.Rank), m.Bg, card.HighlightNone))
				}
			}

			comp := lipgloss.NewCompositor(
				IMap(meldLayers, func(idx int, l *lipgloss.Layer) *lipgloss.Layer {
					return l.Y(idx * 3).Z(idx)
				})...,
			)

			render := comp.Render()
			if m.Focus == FocusMelds && i == m.MeldCursor {
				render = selectedStyle.Render(render)
			}
			meldRenders = append(meldRenders, render)
		}

		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, meldRenders...) + "\n\n")
	}

	// Player's hand
	handLabel := infoStyle.Render(fmt.Sprintf("Your Hand (%d cards):", len(m.Snapshot.YourHand)))
	if m.SelectMode && len(m.Selected) > 0 {
		handLabel += "  " + turnStyle.Render(fmt.Sprintf("%d selected — [enter] to meld, [esc] to cancel", len(m.Selected)))
	}
	sb.WriteString(handLabel + "\n")

	if len(m.Snapshot.YourHand) > 0 {
		const raiseAmount = 2 // How much to raise cursor/selected cards

		var handLayers []*lipgloss.Layer
		for i, c := range m.Snapshot.YourHand {
			suit := convertSuit(c.Suit)
			isCursor := m.Focus == FocusHand && i == m.CursorPos
			isSelected := m.Selected[i]

			// Determine highlight style
			highlight := card.HighlightNone
			if isCursor {
				highlight = card.HighlightCursor
			} else if isSelected {
				highlight = card.HighlightSelected
			}

			isActive := isCursor || isSelected

			if i == len(m.Snapshot.YourHand)-1 {
				// Last card renders full
				handLayers = append(handLayers, card.CardLayer(suit, strconv.Itoa(c.Rank), m.Bg, highlight))
			} else {
				handLayers = append(handLayers, card.HalfCardLayer(isActive, suit, strconv.Itoa(c.Rank), m.Bg, highlight))
			}
		}

		// Compose hand with Y-offsets: cursor/selected cards are raised
		hand := lipgloss.NewCompositor(
			IMap(handLayers, func(i int, l *lipgloss.Layer) *lipgloss.Layer {
				y := raiseAmount // Default: cards sit at baseline
				if m.Focus == FocusHand && i == m.CursorPos {
					y = 0 // Cursor card raised to top
				} else if m.Selected[i] {
					y = 1 // Selected cards raised slightly less
				}
				return l.X(i * 3).Y(y).Z(i)
			})...,
		)

		handRender := hand.Render()
		if m.Focus == FocusHand {
			handRender = selectedStyle.Render(handRender)
		}

		// Center the hand
		leftPad := lipgloss.NewStyle().Width(max(0, (m.Width-hand.Bounds().Dx())/2))
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, leftPad.Render(" "), handRender))

		// Cursor indicator row below the hand
		var indicators strings.Builder
		charsPerCard := 3 // Matches the X spacing in the compositor
		for i := range m.Snapshot.YourHand {
			marker := strings.Repeat(" ", charsPerCard)
			if m.Focus == FocusHand && i == m.CursorPos {
				marker = " ▲ "
			} else if m.Selected[i] {
				marker = " ✓ "
			}
			indicators.WriteString(marker)
		}
		indLine := indicators.String()
		// Pad indicator line to roughly center it under the hand
		indPadWidth := max(0, (m.Width-len([]rune(indLine)))/2)
		sb.WriteString("\n" + strings.Repeat(" ", indPadWidth) + indLine)
	}

	// Status message
	if m.StatusMsg != "" {
		sb.WriteString("\n\n" + errorStyle.Render(m.StatusMsg))
	}

	// Help bar
	sb.WriteString("\n\n" + helpStyle.Render("[tab] focus  [←→] navigate  [space] select  [enter] meld/confirm  [d] draw  [x] discard  [esc] cancel  [q] quit"))

	v := tea.NewView(sb.String())
	v.AltScreen = true
	return v
}

func (m Model) viewGameOver() tea.View {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("🎉 GAME OVER! 🎉") + "\n\n")

	if m.Snapshot.Winner == m.PlayerID {
		sb.WriteString(turnStyle.Render("★ YOU WON! ★") + "\n\n")
	} else {
		winner := m.Snapshot.Winner[:min(8, len(m.Snapshot.Winner))]
		sb.WriteString(infoStyle.Render(fmt.Sprintf("%s wins!", winner)) + "\n\n")
	}

	sb.WriteString("Final Scores:\n")
	for _, p := range m.Snapshot.Players {
		marker := "  "
		if p.ID == m.Snapshot.Winner {
			marker = "★ "
		}
		name := p.Name[:min(8, len(p.Name))]
		you := ""
		if p.IsYou {
			you = " (you)"
		}
		sb.WriteString(fmt.Sprintf("  %s%s: %d points%s\n", marker, name, p.Score, you))
	}

	sb.WriteString(helpStyle.Render("\n\nPress [q] to quit"))

	content := lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, sb.String())
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// --- Helpers ---

// convertSuit maps game.Suit to card.Suit for rendering.
func convertSuit(s game.Suit) card.Suit {
	switch s {
	case game.Spades:
		return card.Spade
	case game.Hearts:
		return card.Heart
	case game.Diamonds:
		return card.Diamond
	case game.Clubs:
		return card.Club
	}
	return card.Spade
}

// IMap applies fn to each element with index, returning the mapped slice.
func IMap[T, V any](ts []T, fn func(int, T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(i, t)
	}
	return result
}
