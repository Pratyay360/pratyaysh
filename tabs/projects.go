package tabs

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type project struct {
	name        string
	description string
	stack       string
}

type Projects struct {
	width    int
	selected int
	items    []project
}

func NewProjects(width int) Projects {
	return Projects{
		width: width,
		items: []project{
			{name: "Terminal Portfolio", description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.", stack: "Go / Bubble Tea / Wish"},
			{name: "Project Dolor", description: "Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", stack: "Go / SQLite / Docker"},
			{name: "Project Amet", description: "Ut enim ad minim veniam, quis nostrud exercitation ullamco.", stack: "TypeScript / React / PostgreSQL"},
		},
	}
}

func (p Projects) Init() tea.Cmd { return nil }

func (p Projects) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			p.selected = (p.selected - 1 + len(p.items)) % len(p.items)
		case "down", "j":
			p.selected = (p.selected + 1) % len(p.items)
		}
	}
	return p, nil
}

func (p Projects) View() tea.View {
	accent := lipgloss.Color("205")
	muted := lipgloss.Color("243")
	width := max(20, min(70, p.width-12))
	rows := make([]string, len(p.items))

	for i, item := range p.items {
		marker := "  "
		nameStyle := lipgloss.NewStyle().Bold(true)
		if i == p.selected {
			marker = "> "
			nameStyle = nameStyle.Foreground(accent)
		}
		rows[i] = fmt.Sprintf("%s%02d  %s\n%s\n%s",
			marker,
			i+1,
			nameStyle.Render(item.name),
			lipgloss.NewStyle().PaddingLeft(6).Width(width).Render(item.description),
			lipgloss.NewStyle().PaddingLeft(6).Foreground(muted).Render(item.stack),
		)
	}

	help := lipgloss.NewStyle().Foreground(muted).Render("Up/Down or j/k: select project")
	return tea.NewView(strings.Join(rows, "\n\n") + "\n\n" + help)
}
