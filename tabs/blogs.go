package tabs

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type article struct {
	title   string
	summary string
	date    string
}

type Blogs struct {
	width    int
	selected int
	articles []article
}

func NewBlogs(width int) Blogs {
	return Blogs{
		width: width,
		articles: []article{
			{title: "Building terminal interfaces in Go", summary: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.", date: "2026-07-30"},
			{title: "Small programs, clear ideas", summary: "Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", date: "2026-06-12"},
			{title: "Learning in public", summary: "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.", date: "2026-05-04"},
		},
	}
}

func (b Blogs) Init() tea.Cmd { return nil }

func (b Blogs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			b.selected = (b.selected - 1 + len(b.articles)) % len(b.articles)
		case "down", "j":
			b.selected = (b.selected + 1) % len(b.articles)
		}
	}
	return b, nil
}

func (b Blogs) View() tea.View {
	accent := lipgloss.Color("205")
	muted := lipgloss.Color("243")
	width := max(20, min(70, b.width-12))
	rows := make([]string, len(b.articles))

	for i, item := range b.articles {
		marker := "  "
		titleStyle := lipgloss.NewStyle().Bold(true)
		if i == b.selected {
			marker = "> "
			titleStyle = titleStyle.Foreground(accent)
		}
		rows[i] = fmt.Sprintf("%s%s  %s\n%s",
			marker,
			lipgloss.NewStyle().Foreground(muted).Render(item.date),
			titleStyle.Render(item.title),
			lipgloss.NewStyle().PaddingLeft(2).Width(width).Render(item.summary),
		)
	}

	help := lipgloss.NewStyle().Foreground(muted).Render("Up/Down or j/k: select article")
	return tea.NewView(strings.Join(rows, "\n\n") + "\n\n" + help)
}
