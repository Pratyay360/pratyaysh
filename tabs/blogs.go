package tabs

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Pratyay360/pratyaysh/libs"
)

type blogsLoadedMsg struct {
	articles []libs.BlogArticle
	err      error
}

type blogsPreviewMsg struct {
	content string
	err     error
	url     string
}

type Blogs struct {
	width    int
	selected int
	articles []libs.BlogArticle
	loading  bool
	err      error

	// preview of selected markdown
	preview        string
	previewURL     string
	previewLoading bool
	previewErr     error
}

func NewBlogs(width int) Blogs {
	return Blogs{width: width, loading: true}
}

func (b Blogs) Init() tea.Cmd {
	return fetchBlogs
}

func fetchBlogs() tea.Msg {
	arts, err := libs.QueryBlogs(context.Background())
	return blogsLoadedMsg{articles: arts, err: err}
}

func fetchPreview(url string) tea.Cmd {
	return func() tea.Msg {
		c, err := libs.Fetch(context.Background(), url)
		if err != nil {
			return blogsPreviewMsg{err: err, url: url}
		}

		if len(c) > 1200 {
			c = c[:1200] + "\n\n… (truncated)"
		}
		return blogsPreviewMsg{content: c, url: url}
	}
}

func (b Blogs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width

	case blogsLoadedMsg:
		b.loading = false
		b.err = msg.err
		b.articles = msg.articles
		b.selected = 0
		if len(b.articles) > 0 {
			b.previewLoading = true
			b.previewURL = b.articles[0].URL
			return b, fetchPreview(b.articles[0].URL)
		}

	case blogsPreviewMsg:
		if b.previewURL != "" && msg.url != b.previewURL {
			return b, nil
		}
		b.previewLoading = false
		b.previewErr = msg.err
		b.preview = msg.content
		b.previewURL = msg.url

	case tea.KeyPressMsg:
		if b.loading {
			return b, nil
		}
		if b.err != nil {
			if msg.String() == "r" {
				b.loading, b.err = true, nil
				return b, fetchBlogs
			}
			return b, nil
		}
		if len(b.articles) == 0 {
			if msg.String() == "r" {
				b.loading = true
				return b, fetchBlogs
			}
			return b, nil
		}
		switch msg.String() {
		case "up", "k":
			b.selected = (b.selected - 1 + len(b.articles)) % len(b.articles)
			b.previewLoading = true
			b.previewErr = nil
			b.previewURL = b.articles[b.selected].URL
			return b, fetchPreview(b.articles[b.selected].URL)
		case "down", "j":
			b.selected = (b.selected + 1) % len(b.articles)
			b.previewLoading = true
			b.previewErr = nil
			b.previewURL = b.articles[b.selected].URL
			return b, fetchPreview(b.articles[b.selected].URL)
		case "r":
			b.loading, b.err, b.articles = true, nil, nil
			b.preview = ""
			return b, fetchBlogs
		case "enter":
			b.previewLoading = true
			b.previewURL = b.articles[b.selected].URL
			return b, fetchPreview(b.articles[b.selected].URL)
		}
	}
	return b, nil
}

func (b Blogs) View() tea.View {
	width := contentWidth(b.width)

	switch {
	case b.loading:
		return tea.NewView(mutedStyle.Render("Fetching articles from blogs_md…") + "\n---FOOTER---\n" + mutedStyle.Render("g/G scroll • tab to switch"))
	case b.err != nil:
		mainContent := strings.Join([]string{
			errorStyle.Render("Could not load blogs."),
			lipgloss.NewStyle().Width(width).Foreground(Muted).Render(b.err.Error()),
			"",
			mutedStyle.Render("r: retry"),
		}, "\n")
		return tea.NewView(mainContent + "\n---FOOTER---\n" + mutedStyle.Render("g/G scroll • tab to switch"))
	case len(b.articles) == 0:
		mainContent := strings.Join([]string{
			mutedStyle.Render("Nothing published yet."),
			"",
			mutedStyle.Render("Add markdown files to github.com/pratyay360/blogs_md"),
			mutedStyle.Render("r: retry"),
		}, "\n")
		return tea.NewView(mainContent + "\n---FOOTER---\n" + mutedStyle.Render("g/G scroll • tab to switch"))
	}

	rows := make([]string, len(b.articles))
	for i, item := range b.articles {
		marker, titleStyle := "  ", boldStyle
		if i == b.selected {
			marker, titleStyle = "> ", selectedStyle
		}
		// Derive display date from path if needed, fallback to empty
		meta := mutedStyle.Render(item.Path)
		rows[i] = fmt.Sprintf("%s%s\n%s",
			marker,
			libs.Link(item.URL, titleStyle.Render(item.Title)),
			indentStyle.Width(width).Render(meta),
		)
	}

	previewBox := ""
	if b.previewLoading {
		previewBox = mutedStyle.Render("Loading preview…")
	} else if b.previewErr != nil {
		previewBox = errorStyle.Render("Preview failed: " + b.previewErr.Error())
	} else if b.preview != "" {

		renderWidth := width - 4
		if renderWidth < 20 {
			renderWidth = width
		}
		if rendered, err := libs.RenderMarkdown(b.preview, renderWidth); err == nil {
			previewBox = lipgloss.NewStyle().
				Width(width - 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Subtle).
				Margin(1).
				Render(strings.TrimSpace(rendered))
		} else {
			// fallback to raw preview on render error
			previewBox = lipgloss.NewStyle().
				Width(width - 2).
				Foreground(Muted).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Subtle).
				Margin(1).
				Render(b.preview)
		}
	}

	mainContent := strings.Join(rows, "\n\n")
	if previewBox != "" {
		mainContent += "\n\n" + previewBox
	}

	help := mutedStyle.Render("Up/Down j/k: select • enter: reload preview • r: refresh • g/G scroll • ctrl+click link to open")
	return tea.NewView(mainContent + "\n---FOOTER---\n" + help)
}
