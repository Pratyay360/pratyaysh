package tabs

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type contact struct {
	label string
	url   string
}

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

func (a About) View() tea.View {
	accent := lipgloss.Color("205")
	muted := lipgloss.Color("243")
	width := max(20, min(70, a.width-12))

	heading := lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Hello, I am Pratyay.")
	bio := lipgloss.NewStyle().Width(width).Render(
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod " +
			"tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
	)

	links := make([]string, len(a.contacts))
	for i, item := range a.contacts {
		marker := "  "
		style := lipgloss.NewStyle().Foreground(muted)
		if i == a.selected {
			marker = "> "
			style = style.Bold(true).Foreground(accent)
		}
		links[i] = style.Render(fmt.Sprintf("%s%-9s %s", marker, item.label, item.url))
	}

	help := lipgloss.NewStyle().Foreground(muted).Render("Up/Down or j/k: focus contact")
	return tea.NewView(strings.Join([]string{heading, "", bio, "", "Contact", strings.Join(links, "\n"), "", help}, "\n"))
}
