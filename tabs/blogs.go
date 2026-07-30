package tabs

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Pratyay360/pratyaysh/libs"
)

type article struct {
	title   string
	summary string
	date    string
	url     string
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
			{
				title:   "",
				summary: "",
				date:    "",
				url:     "",
			},
		},
	}
}

func (b Blogs) Init() tea.Cmd { return nil }

func (b Blogs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
	case tea.KeyPressMsg:
		if len(b.articles) == 0 {
			return b, nil
		}
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
	width := contentWidth(b.width)

	if len(b.articles) == 0 {
		return tea.NewView(mutedStyle.Render("Nothing published yet."))
	}

	rows := make([]string, len(b.articles))
	for i, item := range b.articles {
		marker, titleStyle := "  ", boldStyle
		if i == b.selected {
			marker, titleStyle = "> ", selectedStyle
		}
		rows[i] = fmt.Sprintf("%s%s  %s\n%s",
			marker,
			mutedStyle.Render(item.date),
			libs.Link(item.url, titleStyle.Render(item.title)),
			indentStyle.PaddingLeft(2).Width(width).Render(item.summary),
		)
	}

	help := mutedStyle.Render("Up/Down or j/k: select article")
	return tea.NewView(strings.Join(rows, "\n\n") + "\n\n" + help)
}
