package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/ssh"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lukasmwerner/super-secure-rummy/client"
	"github.com/lukasmwerner/super-secure-rummy/game"
)

type RoomManager struct {
	mu   sync.Mutex
	hubs []*game.Hub
}

func (rm *RoomManager) GetHub(playerID string) *game.Hub {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Try to find an existing hub that allows this player to join
	for _, h := range rm.hubs {
		if h.CanJoin(playerID) {
			return h
		}
	}

	// No available hub found, create a new one
	log.Info("All lobbies full, creating new game room")
	h := game.NewHub(2, 4)
	go h.Run()
	rm.hubs = append(rm.hubs, h)
	return h
}

var rm = &RoomManager{}

func main() {
	host := os.Getenv("RUMMY_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("RUMMY_PORT")
	if port == "" {
		port = "23234"
	}

	// We now use the global RoomManager instead of a single Hub

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true // Accept all public keys
		}),
		ssh.AllocatePty(),
		wish.WithMiddleware(
			// Use MiddlewareWithProgramHandler so we get access to *tea.Program
			// for wiring hub→client snapshot delivery via p.Send()
			bubbletea.MiddlewareWithProgramHandler(makeProgramHandler()),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("SSH server started", "host", host, "port", port)

	go func() {
		err = s.ListenAndServe()
		if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	log.Info("Shutdown triggered", "reason", <-done)
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = s.Shutdown(ctx)
	if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

// makeProgramHandler returns a ProgramHandler that creates a tea.Program
// and wires the PlayerConn.Send to p.Send before the program runs.
func makeProgramHandler() bubbletea.ProgramHandler {
	return func(s ssh.Session) *tea.Program {
		pty, _, _ := s.Pty()
		playerID := fingerprint(s.PublicKey())

		// Assign player to an open hub (or a new one)
		hub := rm.GetHub(playerID)

		log.Info("Player connected to hub", "id", playerID[:8], "term", pty.Term)

		conn := &game.PlayerConn{
			ID:   playerID,
			Name: playerID[:8],
		}

		model := client.Model{
			PlayerID: playerID,
			Width:    pty.Window.Width,
			Height:   pty.Window.Height,
			Bg:       "dark",
			Hub:      hub,
			Conn:     conn,
			Focus:    client.FocusHand,
			Selected: make(map[int]bool),
		}

		// Create the program with wish's I/O options
		p := tea.NewProgram(model, bubbletea.MakeOptions(s)...)

		// Wire the Send function — tea.Program.Send is goroutine-safe.
		// The Hub will call this to inject GameSnapshotMsg into this player's
		// BubbleTea event loop.
		conn.Send = func(msg interface{}) {
			p.Send(msg)
		}

		// Ensure the player is unregistered if the SSH session drops or ends
		go func() {
			<-s.Context().Done()
			hub.Unregister <- playerID
		}()

		return p
	}
}

// fingerprint creates a hex fingerprint of an SSH public key.
func fingerprint(key ssh.PublicKey) string {
	hash := sha256.Sum256(key.Marshal())
	return hex.EncodeToString(hash[:])
}
