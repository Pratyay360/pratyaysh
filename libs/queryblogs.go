package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
)

// BlogArticle represents a markdown file living in github.com/pratyay360/blogs_md
type BlogArticle struct {
	Path    string `json:"path"`
	URL     string `json:"-"` // raw.githubusercontent.com URL
	Title   string `json:"-"`
	Summary string `json:"-"`
	Date    string `json:"-"`
	Type    string `json:"type"` // "blob" | "tree"
}

const contentURL = "https://raw.githubusercontent.com/pratyay360/blogs_md/main/"

// QueryBlogs fetches the full git tree and returns only markdown blobs.
func QueryBlogs(ctx context.Context) ([]BlogArticle, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/pratyay360/blogs_md/git/trees/main?recursive=1", nil)
	if err != nil {
		return nil, fmt.Errorf("blogs: build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if tok := githubToken(); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("blogs: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("blogs: %s %s", resp.Status, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("blogs: read body: %w", err)
	}

	var result struct {
		Tree []BlogArticle `json:"tree"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("blogs: decode: %w", err)
	}

	var out []BlogArticle
	for _, a := range result.Tree {
		if a.Type != "blob" {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(a.Path), ".md") {
			continue
		}
		// normalize title from filename
		base := path.Base(a.Path)
		title := strings.TrimSuffix(base, path.Ext(base))
		title = strings.ReplaceAll(title, "-", " ")
		title = strings.ReplaceAll(title, "_", " ")

		a.Title = title
		a.URL = contentURL + a.Path // preserve subfolder structure
		// keep Path for fallback display
		out = append(out, a)
	}

	return out, nil
}
