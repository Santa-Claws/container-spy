package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorGray  = lipgloss.Color("8")
	colorRed   = lipgloss.Color("9")
	colorGreen = lipgloss.Color("10")

	styleBar = lipgloss.NewStyle().Foreground(colorGray)
	styleErr = lipgloss.NewStyle().Foreground(colorRed)
	styleOK  = lipgloss.NewStyle().Foreground(colorGreen)
)

// StatusBar renders a one-line status bar at the bottom of the screen.
type StatusBar struct {
	width       int
	lastRefresh time.Time
	errCount    int
	mode        string
	hint        string
}

// NewStatusBar creates a StatusBar.
func NewStatusBar(mode string) StatusBar {
	return StatusBar{mode: mode}
}

// SetWidth updates the bar width for rendering.
func (s *StatusBar) SetWidth(w int) { s.width = w }

// Update refreshes the bar with current state data.
func (s *StatusBar) Update(lastRefresh time.Time, errCount int, hint string) {
	s.lastRefresh = lastRefresh
	s.errCount = errCount
	s.hint = hint
}

// View renders the status bar string.
func (s StatusBar) View() string {
	left := fmt.Sprintf(" container-spy [%s]", s.mode)

	refreshStr := "never"
	if !s.lastRefresh.IsZero() {
		refreshStr = s.lastRefresh.Format("15:04:05")
	}

	errStr := ""
	if s.errCount > 0 {
		errStr = styleErr.Render(fmt.Sprintf("  %d error(s)", s.errCount))
	} else if !s.lastRefresh.IsZero() {
		errStr = styleOK.Render("  ✓")
	}

	right := fmt.Sprintf("refreshed: %s%s  %s ", refreshStr, errStr, s.hint)

	gap := s.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return styleBar.Render(left) +
		lipgloss.NewStyle().Width(gap).Render("") +
		styleBar.Render(right)
}
