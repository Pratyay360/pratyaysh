package tabs

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Pratyay360/pratyaysh/libs"
)

const githubUser = "Pratyay360"
type reposLoadedMsg struct {
	repos []libs.Repo
	err   error
}

type Projects struct {
	width    int
	selected int
	repos    []libs.Repo
	loading  bool
	err      error
}

func NewProjects(width int) Projects {
	return Projects{width: width, loading: true}
}

// Init kicks off the repo fetch off the render path.
func (p Projects) Init() tea.Cmd {
	return fetchRepos
}

func fetchRepos() tea.Msg {
	repos, err := libs.GetRepos(context.Background(), githubUser)
	return reposLoadedMsg{repos: repos, err: err}
}

func (p Projects) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
	case reposLoadedMsg:
		p.loading = false
		p.repos = msg.repos
		p.err = msg.err
		p.selected = 0
	case tea.KeyPressMsg:
		if len(p.repos) == 0 {
			if msg.String() == "r" {
				p.loading, p.err = true, nil
				return p, fetchRepos
			}
			return p, nil
		}
		switch msg.String() {
		case "up", "k":
			p.selected = (p.selected - 1 + len(p.repos)) % len(p.repos)
		case "down", "j":
			p.selected = (p.selected + 1) % len(p.repos)
		case "r":
			p.loading, p.err, p.repos = true, nil, nil
			return p, fetchRepos
		}
	}
	return p, nil
}

func (p Projects) View() tea.View {
	width := contentWidth(p.width)

	switch {
	case p.loading:
		return tea.NewView(mutedStyle.Render("Fetching repositories from GitHub..."))
	case p.err != nil:
		return tea.NewView(strings.Join([]string{
			errorStyle.Render("Could not load projects."),
			lipgloss.NewStyle().Width(width).Foreground(Muted).Render(p.err.Error()),
			"",
			mutedStyle.Render("r: retry"),
		}, "\n"))
	case len(p.repos) == 0:
		return tea.NewView(mutedStyle.Render("No public repositories to show.\n\nr: retry"))
	}

	rows := make([]string, len(p.repos))
	for i, repo := range p.repos {
		marker, nameStyle := "  ", boldStyle
		if i == p.selected {
			marker, nameStyle = "> ", selectedStyle
		}

		description := repo.Description
		if description == "" {
			description = "No description provided."
		}

		meta := repo.Language
		if meta == "" {
			meta = "unknown"
		}
		if repo.Stars > 0 {
			meta = fmt.Sprintf("%s  *%d", meta, repo.Stars)
		}

		rows[i] = fmt.Sprintf("%s%02d  %s\n%s\n%s",
			marker,
			i+1,
			libs.Link(repo.URL, nameStyle.Render(repo.Name)),
			indentStyle.Width(width).Render(description),
			indentStyle.Foreground(Muted).Render(meta),
		)
	}

	help := mutedStyle.Render("Up/Down or j/k: select | r: refresh")
	return tea.NewView(strings.Join(rows, "\n\n") + "\n\n" + help)
}
