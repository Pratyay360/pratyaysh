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

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	wishtea "charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/Pratyay360/pratyaysh/tabs"
	"github.com/charmbracelet/ssh"
)

const (
	host = "0.0.0.0"
	port = "2222"
)

var tabNames = []string{"About", "Projects", "Blogs"}

func main() {
	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/ssh"),
		wish.WithMiddleware(
			logging.Middleware(),
			wishtea.Middleware(teaHandler),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("Could not create SSH server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("SSH server stopped unexpectedly", "error", err)
			done <- syscall.SIGTERM
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop SSH server", "error", err)
	}
}

type model struct {
	width     int
	height    int
	activeTab int
	tabs      []tea.Model
}

func teaHandler(session ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := session.Pty()
	return newModel(pty.Window.Width, pty.Window.Height), wishtea.MakeOptions(session)
}

func newModel(width, height int) model {
	return model{
		width:  width,
		height: height,
		tabs: []tea.Model{
			tabs.NewAbout(width),
			tabs.NewProjects(width),
			tabs.NewBlogs(width),
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
		for i, tab := range m.tabs {
			m.tabs[i], _ = tab.Update(msg)
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % len(tabNames)
		case "shift+tab", "left", "h":
			m.activeTab = (m.activeTab - 1 + len(tabNames)) % len(tabNames)
		case "1", "2", "3":
			m.activeTab = int(msg.String()[0] - '1')
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

	title := lipgloss.NewStyle().Bold(true).Foreground(accent).Render("PRATYAY MUSTAFI")
	subtitle := lipgloss.NewStyle().Foreground(muted).Render("developer / builder / curious human")

	labels := make([]string, len(tabNames))
	for i, name := range tabNames {
		style := lipgloss.NewStyle().Padding(0, 1).Foreground(muted)
		if i == m.activeTab {
			style = style.Bold(true).Foreground(lipgloss.Color("230")).Background(accent)
		}
		labels[i] = style.Render(fmt.Sprintf("%d %s", i+1, name))
	}

	content := m.tabs[m.activeTab].View().Content
	panel := lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Render(content)

	help := lipgloss.NewStyle().Foreground(muted).Render("Tab/Shift+Tab or Left/Right: switch | 1-3: jump | q: quit")
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

	view := tea.NewView(lipgloss.NewStyle().Padding(1, 2).Render(body))
	view.AltScreen = true
	return view
}
