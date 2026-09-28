package libs

import (
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func githubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}
