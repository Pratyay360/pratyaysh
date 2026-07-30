package libs

import "github.com/charmbracelet/x/ansi"

func Link(url, label string) string {
	if url == "" {
		return label
	}
	if label == "" {
		label = url
	}
	return ansi.SetHyperlink(url) + label + ansi.ResetHyperlink()
}
