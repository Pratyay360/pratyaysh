package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
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
	"github.com/Pratyay360/pratyaysh/about"
	"github.com/Pratyay360/pratyaysh/projects"
	"github.com/Pratyay360/pratyaysh/tabs"
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
	// Railway / PaaS healthcheck: expose HTTP on $PORT so the platform sees the service as healthy.
	// The SSH TUI still listens on 2222; this is just for the orchestrator.
	if p := os.Getenv("PORT"); p != "" {
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, _ = w.Write([]byte("pratyaysh ssh: ssh ssh.pratyay.qzz.io\n"))
			})
			mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})
			log.Info("Starting health server", "port", p)
			if err := http.ListenAndServe(net.JoinHostPort("", p), mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error("Health server error", "error", err)
			}
		}()
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "about":
			about.About()
			return
		case "projects":
			projects.ListProjects()
			return
		case "help", "-h", "--help":
			fmt.Println("pratyaysh - interactive terminal resume & portfolio")
			fmt.Println("\nUsage:")
			fmt.Println("  pratyaysh            Start SSH server (default)")
			fmt.Println("  pratyaysh serve      Start SSH server explicitly")
			fmt.Println("  pratyaysh about      Print about information")
			fmt.Println("  pratyaysh projects   List projects")
			return
		case "serve":
			// Proceed to start SSH server
		}
	}

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
			return m, nil

		case "shift+tab", "left", "h":
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = len(m.tabs) - 1
			}
			return m, nil

		case "1", "2", "3", "4":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.tabs) {
				m.activeTab = idx
			}
			return m, nil

		default:
			var cmd tea.Cmd
			m.tabs[m.activeTab], cmd = m.tabs[m.activeTab].Update(msg)
			return m, cmd
		}

	default:
		// Broadcast async messages (reposLoadedMsg, blogsLoadedMsg, preview, etc.)
		// to all tabs. Forwarding only to the active tab would drop results
		// that arrive while the user is on another tab, leaving loading stuck.
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

	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, labels...)

	panelContent := m.tabs[m.activeTab].View().Content
	panel := lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Render(panelContent)

	help := lipgloss.NewStyle().
		Foreground(muted).
		Render("Tab/Shift+Tab • ←/→ h/l • 1-4 switch • q/esc quit   •   ↑/↓ j/k inside tabs • r refresh")

	body := strings.Join([]string{
		title,
		subtitle,
		"",
		tabsRow,
		"",
		panel,
		"",
		help,
	}, "\n")

	// Center vertically when we know the terminal height
	rendered := lipgloss.NewStyle().
		Padding(1, 2).
		Render(body)

	if m.height > 0 {
		// Place helper would be ideal, but manual height guard keeps small terminals readable.
		// Don't use PlaceVertical – it truncates. Just ensure we don't overflow.
		_ = m.height
	}

	return tea.NewView(rendered)
}
