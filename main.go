package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "charm.land/bubbles/v2"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/ssh"
	_ "github.com/joho/godotenv/autoload"
)

type sessionState uint

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func main() {
	host := os.Getenv("RUMMY_HOST")
	if len(host) == 0 {
		host = "localhost"
	}
	port := os.Getenv("RUMMY_PORT")
	if len(port) == 0 {
		port = "23234"
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
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

	t := term_model{
		term:            pty.Term,
		width:           pty.Window.Width,
		height:          pty.Window.Height,
		text_style:      lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		quit_text_style: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		help:            false,
	}

	return t, []tea.ProgramOption{}
}

type term_model struct {
	term            string
	width           int
	height          int
	bg              string
	color_profile   string
	text_style      lipgloss.Style
	quit_text_style lipgloss.Style
	help            bool
	//state           sessionState
}

func (m term_model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
	)
}

func (m term_model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.ColorProfileMsg:
		m.color_profile = msg.String()
	case tea.BackgroundColorMsg:
		if msg.IsDark() {
			m.bg = "dark"
		} else {
			m.bg = "light"
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if !m.help {
				return m, tea.Quit
			}
			m.help = false
		case "?":
			m.help = true
		}
	}
	return m, nil
}

func (m term_model) View() tea.View {

	var st strings.Builder
	//model := m.currentFocusedModel()
	if m.help {
		vp := viewport.New()
		vp.SetWidth(m.width / 2)
		vp.SetHeight(m.height / 2)
		vp.Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			PaddingRight(2)
		vp.SetContent(fmt.Sprintf("Welcome to Secure Rummy!\nHere are the basic commands:\n ...\n"))
		//vp.View()

		v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, vp.View()))
		v.AltScreen = true

		return v
	}

	s := fmt.Sprintf("Your term is %s\nYour window size is %dx%d\nBackground: %s\nColor Profile: %s", m.term, m.width, m.height, m.bg, m.color_profile)
	st.WriteString(helpStyle.Render(fmt.Sprintf("\n?: help, q: exit\n")))
	v := tea.NewView(m.text_style.Render(s) + "\n\n" + m.quit_text_style.Render(st.String()))
	v.AltScreen = true
	return v
}
