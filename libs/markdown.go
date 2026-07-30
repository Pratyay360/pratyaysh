package libs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"charm.land/glamour/v2"
)

var httpClient = &http.Client{}

func Fetch(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(body), nil
}

func RenderMarkdown(source string, width int) (string, error) {
	if width < 20 {
		width = 20
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
	)
	if err != nil {
		return "", fmt.Errorf("new renderer: %w", err)
	}
	out, err := r.Render(source)
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return out, nil
}

func Render(ctx context.Context, url string, width int) (string, error) {
	source, err := Fetch(ctx, url)
	if err != nil {
		return "", err
	}
	return RenderMarkdown(source, width)
}
