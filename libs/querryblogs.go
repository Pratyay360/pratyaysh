package libs

import (
	"encoding/json"
	"io"
	"net/http"
)

type article struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func QueryBlogs() ([]article, error) {
	resp, err := http.Get("https://api.github.com/repos/pratyay360/blogs_md/git/trees/main?recursive=1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tree []article
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Tree, nil
}
