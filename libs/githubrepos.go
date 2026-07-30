package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"
)

// userAgent is required for polite GitHub API access; unauthenticated requests
// without one are rate limited far more aggressively.
const userAgent = "pratyaysh-tui (+https://github.com/Pratyay360/pratyaysh)"

const (
	reposPerPage = 100
	maxRepoPages = 3
)

// Repo is a trimmed-down GitHub repository record.
type Repo struct {
	Name        string    `json:"name"`
	URL         string    `json:"html_url"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Stars       int       `json:"stargazers_count"`
	Fork        bool      `json:"fork"`
	Archived    bool      `json:"archived"`
	PushedAt    time.Time `json:"pushed_at"`
	Topics      []string  `json:"topics"`
}

// GetRepos lists the public, non-fork, non-archived repositories for username,
// most-starred first (ties broken by most recently pushed).
func GetRepos(ctx context.Context, username string) ([]Repo, error) {
	if username == "" {
		return nil, fmt.Errorf("github: empty username")
	}

	var all []Repo
	for page := 1; page <= maxRepoPages; page++ {
		batch, err := fetchRepoPage(ctx, username, page)
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < reposPerPage {
			break
		}
	}

	filtered := all[:0]
	for _, r := range all {
		if r.Fork || r.Archived {
			continue
		}
		filtered = append(filtered, r)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Stars != filtered[j].Stars {
			return filtered[i].Stars > filtered[j].Stars
		}
		return filtered[i].PushedAt.After(filtered[j].PushedAt)
	})
	return filtered, nil
}

func fetchRepoPage(ctx context.Context, username string, page int) ([]Repo, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	endpoint := fmt.Sprintf(
		"https://api.github.com/users/%s/repos?per_page=%d&page=%d&sort=pushed",
		url.PathEscape(username), reposPerPage, page,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("github: build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("github: user %q not found", username)
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, fmt.Errorf("github: rate limited (%s)", resp.Status)
	default:
		return nil, fmt.Errorf("github: unexpected status %s", resp.Status)
	}

	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("github: decode response: %w", err)
	}
	return repos, nil
}
