package tabs

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Pratyay360/pratyaysh/libs"
)

type About struct {
	width    int
	selected int
	contacts []contact
}

func NewAbout(width int) About {
	return About{
		width: width,
		contacts: []contact{
			{label: "GitHub", url: "https://github.com/Pratyay360"},
			{label: "LinkedIn", url: "https://linkedin.com/in/pratyay360"},
			{label: "X", url: "https://x.com/realpratyay"},
			{label: "Codeberg", url: "https://codeberg.org/Pratyay360"},
			{label: "Mastodon", url: "https://mastodon.social/@realpratyay"},
			{label: "Bluesky", url: "https://bsky.app/profile/realpratyay"},
			{label: "Instagram", url: "https://instagram.com/realpratyay"},
			{label: "Facebook", url: "https://facebook.com/pratyaymustafi"},
			{label: "GitLab", url: "https://gitlab.com/pratyay360"},
		},
	}
}

func (a About) Init() tea.Cmd { return nil }

func (a About) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			a.selected = (a.selected - 1 + len(a.contacts)) % len(a.contacts)
		case "down", "j":
			a.selected = (a.selected + 1) % len(a.contacts)
		}
	}
	return a, nil
}

const bioText = "Curious developer navigating the ever-shifting landscape of tech — " +
	"Go, infra, and open-source. I like building small tools that feel good to use. " +
	"This is the terminal version of my personal site. SSH in anytime."

func (a About) View() tea.View {
	width := contentWidth(a.width)

	// Simple ASCII banner — figlet is intentionally avoided here: the
	// figlet-go library panics on some terminals and its API doesn't
	// return a string reliably in this bubbletea v2 setup. A lipgloss
	// banner is more portable.
	banner := lipgloss.NewStyle().
		Bold(true).
		Foreground(Accent).
		Width(width).
		Render("pratyay mustafi  —  hello!")

	bio := lipgloss.NewStyle().
		Width(width).
		Foreground(Muted).
		Render(bioText)

	avail := lipgloss.NewStyle().
		Width(width).
		Foreground(lipgloss.Color("243")).
		Render("Available via SSH • " + libs.Link("https://pratyay.dev", "pratyay.dev"))

	links := make([]string, len(a.contacts))
	for i, item := range a.contacts {
		marker, style := "  ", mutedStyle
		if i == a.selected {
			marker, style = "> ", selectedStyle
		}
		row := fmt.Sprintf("%s%-10s %s", marker, item.label, item.url)
		links[i] = libs.Link(item.url, style.Render(row))
	}

	help := mutedStyle.Render("↑/↓ j/k: focus • ctrl+click link to open • tab to switch sections")
	return tea.NewView(strings.Join([]string{
		banner,
		"",
		bio,
		"",
		avail,
		"",
		boldStyle.Render("Find me on"),
		strings.Join(links, "\n"),
		"", help,
	}, "\n"))
}
