package tabs

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Pratyay360/pratyaysh/libs"
)

type contact struct {
	label string
	url   string
}

type Contact struct {
	width    int
	selected int
	contacts []contact
}

func NewContact(width int) Contact {
	return Contact{
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

func (c Contact) Init() tea.Cmd { return nil }

func (c Contact) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			c.selected = (c.selected - 1 + len(c.contacts)) % len(c.contacts)
		case "down", "j":
			c.selected = (c.selected + 1) % len(c.contacts)
		}
	}
	return c, nil
}

func (c Contact) View() tea.View {
	email := libs.Link("mailto:pratyay@example.com", selectedStyle.Render("pratyay@example.com"))

	links := make([]string, len(c.contacts))
	for i, item := range c.contacts {
		marker, style := "  ", mutedStyle
		if i == c.selected {
			marker, style = "> ", selectedStyle
		}
		row := fmt.Sprintf("%s%-10s %s", marker, item.label, item.url)
		links[i] = libs.Link(item.url, style.Render(row))
	}

	help := mutedStyle.Render("Up/Down j/k: focus • ctrl+click to open • or type \"ssh\" from anywhere")
	return tea.NewView(strings.Join([]string{
		mutedStyle.Render("Prefer email for quickest reply:"),
		email,
		"",
		boldStyle.Render("Elsewhere"),
		strings.Join(links, "\n"),
		"", help,
	}, "\n"))
}
