package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorGreen  = lipgloss.Color("10")
	ColorRed    = lipgloss.Color("9")
	ColorYellow = lipgloss.Color("11")
	ColorBlue   = lipgloss.Color("12")
	ColorGray   = lipgloss.Color("8")
	ColorWhite  = lipgloss.Color("15")
	ColorBg     = lipgloss.Color("0")

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlue).
			Padding(0, 1)

	StyleGroupHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorYellow).
				Padding(0, 0, 0, 1)

	StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorGray).
			Padding(0, 1)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorRed)

	StyleSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(ColorWhite)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorGray).
			Italic(true)

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGray)

	StyleRunning = lipgloss.NewStyle().Foreground(ColorGreen)
	StyleExited  = lipgloss.NewStyle().Foreground(ColorRed)
	StyleOther   = lipgloss.NewStyle().Foreground(ColorYellow)
)

// StateStyle returns the appropriate style for a Docker container state string.
func StateStyle(state string) lipgloss.Style {
	switch state {
	case "running":
		return StyleRunning
	case "exited", "dead":
		return StyleExited
	default:
		return StyleOther
	}
}
