package libs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glow/v3/utils"
)

func Fetch(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("fetch %s: build request: %w", url, err)
	}
	if tok := githubToken(); tok != "" && isGitHubURL(url) {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("fetch %s: %s %s", url, resp.Status, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(body), nil
}

func isGitHubURL(u string) bool {
	return strings.HasPrefix(u, "https://api.github.com") ||
		strings.HasPrefix(u, "https://raw.githubusercontent.com")
}

func RenderMarkdown(source string, width int) (string, error) {
	if width <= 0 {
		width = 80
	}
	cleaned := utils.RemoveFrontmatter([]byte(source))
	r, err := glamour.NewTermRenderer(
		utils.GlamourStyle("auto", false),
		glamour.WithWordWrap(width),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}

	out, err := r.Render(string(cleaned))
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return out, nil
}
