package tabs

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Pratyay360/pratyaysh/libs"
)

const (
	githubUser   = "Pratyay360"
	repoFetchTTL = 8 * time.Second
)

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
	ctx, cancel := context.WithTimeout(context.Background(), repoFetchTTL)
	defer cancel()
	repos, err := libs.GetRepos(ctx, githubUser)
	return reposLoadedMsg{repos: repos, err: err}
}

func (p Projects) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		return p, nil

	case reposLoadedMsg:
		p.loading = false
		p.repos = msg.repos
		p.err = msg.err
		p.selected = 0
		return p, nil

	case tea.KeyPressMsg:
		if p.loading {
			return p, nil
		}
		if p.err != nil || len(p.repos) == 0 {
			if msg.String() == "r" {
				p.loading = true
				p.err = nil
				p.repos = nil
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
		case "g":
			p.selected = 0
		case "G":
			p.selected = len(p.repos) - 1
		}
	}
	return p, nil
}

func (p Projects) View() tea.View {
	width := contentWidth(p.width)

	switch {
	case p.loading:
		return tea.NewView(mutedStyle.Render("Fetching repositories from GitHub…"))
	case p.err != nil:
		return tea.NewView(strings.Join([]string{
			errorStyle.Render("Could not load projects."),
			lipgloss.NewStyle().Width(width).Foreground(Muted).Render(p.err.Error()),
			"",
			mutedStyle.Render("r: retry  •  check GITHUB_TOKEN if rate-limited"),
		}, "\n"))
	case len(p.repos) == 0:
		return tea.NewView(strings.Join([]string{
			mutedStyle.Render("No public repositories to show."),
			"",
			mutedStyle.Render("r: retry"),
		}, "\n"))
	}

	// Keep selection in view — simple windowing for SSH narrow terminals.
	// Show at most 8 entries centered around selected; avoids flooding the panel.
	windowSize := 8
	start, end := 0, len(p.repos)
	if len(p.repos) > windowSize {
		half := windowSize / 2
		start = p.selected - half
		if start < 0 {
			start = 0
		}
		end = start + windowSize
		if end > len(p.repos) {
			end = len(p.repos)
			start = end - windowSize
		}
	}

	rows := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		repo := p.repos[i]
		marker, nameStyle := "  ", boldStyle
		if i == p.selected {
			marker, nameStyle = "> ", selectedStyle
		}

		description := strings.TrimSpace(repo.Description)
		if description == "" {
			description = mutedStyle.Render("No description provided.")
		} else {
			description = indentStyle.Width(width).Render(description)
		}

		metaParts := []string{}
		lang := repo.Language
		if lang == "" {
			lang = "unknown"
		}
		metaParts = append(metaParts, lang)
		if repo.Stars > 0 {
			metaParts = append(metaParts, fmt.Sprintf("★ %d", repo.Stars))
		}
		if !repo.PushedAt.IsZero() {
			metaParts = append(metaParts, repo.PushedAt.Format("2006-01-02"))
		}
		if len(repo.Topics) > 0 {
			// Show up to 3 topics to keep lines short
			topics := repo.Topics
			if len(topics) > 3 {
				topics = topics[:3]
			}
			metaParts = append(metaParts, mutedStyle.Render("#"+strings.Join(topics, " #")))
		}
		meta := strings.Join(metaParts, "  •  ")

		row := fmt.Sprintf("%s%02d  %s\n%s\n%s",
			marker,
			i+1,
			libs.Link(repo.URL, nameStyle.Render(repo.Name)),
			description,
			indentStyle.Foreground(Muted).Render(meta),
		)
		rows = append(rows, row)
	}

	header := ""
	if len(p.repos) > windowSize {
		header = mutedStyle.Render(fmt.Sprintf("%d repos — showing %d–%d  •  %d selected",
			len(p.repos), start+1, end, p.selected+1))
	} else {
		header = mutedStyle.Render(fmt.Sprintf("%d repos  •  sorted by stars → recent push", len(p.repos)))
	}

	help := mutedStyle.Render("↑/↓ j/k: select • g/G: top/bottom • r: refresh • ctrl+click link to open")
	content := strings.Join([]string{
		header,
		"",
		strings.Join(rows, "\n\n"),
		"",
		help,
	}, "\n")
	return tea.NewView(content)
}
