package tabs

import "charm.land/lipgloss/v2"

var (
	Accent = lipgloss.Color("205")
	Muted  = lipgloss.Color("243")
	Subtle = lipgloss.Color("238")
	Bright = lipgloss.Color("230")
	Danger = lipgloss.Color("203")
)

var (
	headingStyle  = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	mutedStyle    = lipgloss.NewStyle().Foreground(Muted)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	boldStyle     = lipgloss.NewStyle().Bold(true)
	errorStyle    = lipgloss.NewStyle().Foreground(Danger)
	indentStyle   = lipgloss.NewStyle().PaddingLeft(6)
)
func contentWidth(width int) int {
	return max(20, min(70, width-12))
}
