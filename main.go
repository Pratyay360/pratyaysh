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

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"charm.land/wish/v2"
	wishtea "charm.land/wish/v2/bubbletea"

	"github.com/Pratyay360/pratyaysh/tabs"
	"github.com/charmbracelet/ssh"
)

const (
	host = "0.0.0.0"
	port = "2222"
)

var tabNames = []string{
	"About",
	"Projects",
	"Blogs",
	"Contact",
}

func main() {
	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/ssh"),
		wish.WithMiddleware(
			wishtea.Middleware(teaHandler),
		),
	)
	if err != nil {
		log.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	log.Info("Starting SSH server", "host", host, "port", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("SSH server stopped", "error", err)
			done <- syscall.SIGTERM
		}
	}()

	<-done

	log.Info("Stopping SSH server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Shutdown error", "error", err)
	}
}

type model struct {
	width     int
	height    int
	activeTab int
	tabs      []tea.Model
}

func teaHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, ok := sess.Pty()

	w, h := 80, 24
	if ok {
		w = pty.Window.Width
		h = pty.Window.Height
	}

	return newModel(w, h), wishtea.MakeOptions(sess)
}

func newModel(width, height int) model {
	return model{
		width:  width,
		height: height,
		tabs: []tea.Model{
			tabs.NewAbout(width),
			tabs.NewProjects(width),
			tabs.NewBlogs(width),
			tabs.NewContact(width),
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		for i := range m.tabs {
			m.tabs[i], _ = m.tabs[i].Update(msg)
		}

	case tea.KeyPressMsg:
		switch msg.String() {

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)

		case "shift+tab", "left", "h":
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = len(m.tabs) - 1
			}

		case "1", "2", "3", "4":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.tabs) {
				m.activeTab = idx
			}

		default:
			var cmd tea.Cmd
			m.tabs[m.activeTab], cmd = m.tabs[m.activeTab].Update(msg)
			return m, cmd
		}

	default:
		var cmd tea.Cmd
		m.tabs[m.activeTab], cmd = m.tabs[m.activeTab].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() tea.View {
	accent := lipgloss.Color("205")
	muted := lipgloss.Color("243")

	panelWidth := min(80, m.width-4)
	if panelWidth < 24 {
		panelWidth = 24
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		Render("PRATYAY MUSTAFI")

	subtitle := lipgloss.NewStyle().
		Foreground(muted).
		Render("developer • curious human")

	labels := make([]string, len(tabNames))

	for i, name := range tabNames {
		style := lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(muted)

		if i == m.activeTab {
			style = style.
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Background(accent)
		}

		labels[i] = style.Render(fmt.Sprintf("%d %s", i+1, name))
	}

	panel := lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Render(m.tabs[m.activeTab].View().Content)

	help := lipgloss.NewStyle().
		Foreground(muted).
		Render("Tab/Shift+Tab • ←/→ • 1-4 • q to quit")

	body := strings.Join([]string{
		title,
		subtitle,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, labels...),
		"",
		panel,
		"",
		help,
	}, "\n")

	return tea.NewView(lipgloss.NewStyle().
		Padding(1, 2).
		Render(body))
}
