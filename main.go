package main

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "charm.land/bubbles/v2"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

var gm = game.NewGameManager()

func main() {

	host := os.Getenv("RUMMY_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("RUMMY_PORT")
	if port == "" {
		port = "23234"
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		}),
		ssh.AllocatePty(),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
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
	if err != nil && errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	pubKey := s.PublicKey()
	pubKeyHex := hex.EncodeToString(pubKey.Marshal())

	// pubBytes := pubKey.Marshal()

	gameState, gameID, updateChan := gm.GetOrCreateGame(pubKeyHex)

	t := client.Model{
		Bg:              "dark",
		PubKey:          pubKeyHex,
		Term:            pty.Term,
		Width:           pty.Window.Width,
		Height:          pty.Window.Height,
		Text_style:      lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		Quit_text_style: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Help:            false,
		State:           gameState,
		UpdateChan:      updateChan,
		GameID:          gameID,
		HandLen:         7,
	}

	return t, []tea.ProgramOption{}
}
