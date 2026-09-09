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

	"charm.land/ssh"
	"github.com/Pratyay360/pratyaysh/tabs"
)

const (
	host = "0.0.0.0"
	port = "22"
)

var tabNames = []string{
	"About",
	"Projects",
	"Blogs",
	"Contact",
}

func main() {
	keyPath := hostKeyPath()
	if _, err := os.Stat(keyPath); err != nil {
		log.Warn("Host key not found", "path", keyPath, "error", err)
		log.Warn("Generate one with: ssh-keygen -t ed25519 -f " + keyPath + " -N ''")
	}

	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(keyPath),
		wish.WithIdleTimeout(10*time.Minute),
		wish.WithMiddleware(
			wishtea.Middleware(teaHandler),
			loggingMiddleware,
		),
	)
	if err != nil {
		log.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	log.Info("Starting SSH server", "host", host, "port", port, "key", keyPath)

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

func hostKeyPath() string {
	if p := os.Getenv("SSH_HOST_KEY"); p != "" {
		return p
	}
	// container mounts to /.ssh, local dev uses .ssh/ssh
	candidates := []string{"/.ssh/ssh", ".ssh/ssh", "ssh"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	// fallback – wish will generate an ephemeral key and log, but we still return something sensible
	return ".ssh/ssh"
}

func loggingMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		start := time.Now()
		pty, _, _ := sess.Pty()
		remote := sess.RemoteAddr().String()
		user := sess.User()
		log.Info("session start", "user", user, "remote", remote, "term", pty.Term)
		next(sess)
		log.Info("session end", "user", user, "remote", remote, "duration", time.Since(start).Round(time.Millisecond).String())
	}
}

type model struct {
	width       int
	height      int
	activeTab   int
	scrollOffset int
	tabs        []tea.Model
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
	var cmds []tea.Cmd
	for _, t := range m.tabs {
		if c := t.Init(); c != nil {
			cmds = append(cmds, c)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		for i := range m.tabs {
			m.tabs[i], _ = m.tabs[i].Update(msg)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			m.scrollOffset = 0
			return m, nil

		case "shift+tab", "left", "h":
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = len(m.tabs) - 1
			}
			m.scrollOffset = 0
			return m, nil

		case "1", "2", "3", "4":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.tabs) {
				m.activeTab = idx
				m.scrollOffset = 0
			}
			return m, nil

		case "g":
			m.scrollOffset = 0
			return m, nil

		case "G":
			// scroll to bottom will be computed in View
			m.scrollOffset = -1
			return m, nil

		default:
			var cmd tea.Cmd
			m.tabs[m.activeTab], cmd = m.tabs[m.activeTab].Update(msg)
			return m, cmd
		}

	default:
		var cmds []tea.Cmd
		for i := range m.tabs {
			var cmd tea.Cmd
			m.tabs[i], cmd = m.tabs[i].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if len(cmds) == 0 {
			return m, nil
		}
		return m, tea.Batch(cmds...)
	}
}

func (m model) View() tea.View {
	accent := lipgloss.Color("205")
	muted := lipgloss.Color("243")

	panelWidth := min(100, m.width-4)
	if panelWidth < 30 {
		panelWidth = max(30, m.width-4)
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		Render("PRATYAY MUSTAFI")

	subtitle := lipgloss.NewStyle().
		Foreground(muted).
		Render("developer • curious human  •  ssh pratyaysh — a TUI resume")

	labels := make([]string, len(tabNames))

	for i, name := range tabNames {
		style := lipgloss.NewStyle().
			MarginRight(2).
			Foreground(muted)

		if i == m.activeTab {
			style = style.
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Background(accent)
		}

		labels[i] = style.Render(fmt.Sprintf("%d %s", i+1, name))
	}

	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, labels...)

	rawContent := m.tabs[m.activeTab].View().Content

	// Split into main scrollable content and fixed footer
	mainContent := rawContent
	footer := ""
	if idx := strings.Index(rawContent, "---FOOTER---"); idx != -1 {
		mainContent = strings.TrimSpace(rawContent[:idx])
		footer = strings.TrimSpace(rawContent[idx+len("---FOOTER---"):])
	}

	// Calculate available height for panel content
	// header: title(1) + subtitle(1) + blank(1) + tabs(1) + blank(1) = 5
	// footer area: blank(1) + help(1) = 2
	// panel borders: top(1) + bottom(1) = 2
	// body margin: top(1) + bottom(1) = 2
	// total overhead = 5 + 2 + 2 + 2 = 11
	overhead := 11
	maxContentHeight := m.height - overhead
	if maxContentHeight < 3 {
		maxContentHeight = 3
	}

	lines := strings.Split(mainContent, "\n")
	totalLines := len(lines)

	// Handle G (scroll to bottom)
	if m.scrollOffset == -1 {
		if totalLines > maxContentHeight {
			m.scrollOffset = totalLines - maxContentHeight
		} else {
			m.scrollOffset = 0
		}
	}

	// Clamp scroll offset
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	if totalLines > maxContentHeight && m.scrollOffset > totalLines-maxContentHeight {
		m.scrollOffset = totalLines - maxContentHeight
	}
	if totalLines <= maxContentHeight {
		m.scrollOffset = 0
	}

	// Slice visible lines
	if totalLines > maxContentHeight {
		lines = lines[m.scrollOffset : m.scrollOffset+maxContentHeight]
	}

	panelContent := strings.Join(lines, "\n")

	// Add scroll indicator
	scrollIndicator := ""
	if totalLines > maxContentHeight {
		scrollIndicator = fmt.Sprintf("  [%d/%d]", m.scrollOffset+maxContentHeight, totalLines)
	}

	panel := lipgloss.NewStyle().
		Width(panelWidth).
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Render(panelContent)

	help := lipgloss.NewStyle().
		Foreground(muted).
		Render("Tab/←→ • 1-4 switch • q quit • g/G scroll • ↑/↓ inside tabs • r refresh" + scrollIndicator)

	parts := []string{
		title,
		subtitle,
		"",
		tabsRow,
		"",
		panel,
	}
	if footer != "" {
		parts = append(parts, "", footer)
	}
	parts = append(parts, "", help)

	body := strings.Join(parts, "\n")

	rendered := lipgloss.NewStyle().
		Margin(1).
		Render(body)

	return tea.NewView(rendered)
}
