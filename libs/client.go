package libs

import (
	"net/http"
	"os"
	"time"
)

// httpClient is shared across libs with sensible timeouts.
// GitHub token (if set) is picked up per-request via Authorization header.
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func githubToken() string {
	// Support both plain env and SOPS-decrypted .env
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}
