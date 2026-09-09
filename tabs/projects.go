package tabs

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Pratyay360/pratyaysh/projects"
)

type Projects struct {
	width     int
	rendered  string
	lastWidth int
}

func NewProjects(width int) Projects {
	p := Projects{
		width: width,
	}
	return p.renderContent()
}

func (p Projects) renderContent() Projects {
	w := contentWidth(p.width)
	rendered, err := projects.RenderMarkdown(w)
	if err != nil {
		p.rendered = projects.ProjectMD
	} else {
		p.rendered = rendered
	}
	p.lastWidth = p.width
	return p
}

func (p Projects) Init() tea.Cmd {
	return nil
}

func (p Projects) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p = p.renderContent()
	case tea.KeyPressMsg:
		if msg.String() == "r" {
			p = p.renderContent()
		}
	}
	return p, nil
}

func (p Projects) View() tea.View {
	width := contentWidth(p.width)
	rendered := p.rendered
	if rendered == "" || p.width != p.lastWidth {
		if r, err := projects.RenderMarkdown(width); err == nil {
			rendered = r
		} else {
			rendered = projects.ProjectMD
		}
	}

	help := mutedStyle.Render("r: reload • g/G scroll • tab to switch")
	content := rendered + "\n---FOOTER---\n" + help

	return tea.NewView(content)
}
