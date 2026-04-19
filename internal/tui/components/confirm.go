package components

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmedMsg is sent when the user confirms (Y).
type ConfirmedMsg struct{ ID string }

// CancelledMsg is sent when the user cancels (N/esc).
type CancelledMsg struct{}

var confirmStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("11")).
	Padding(1, 3)

// Confirm is a simple Y/N overlay dialog.
type Confirm struct {
	prompt  string
	id      string
	visible bool
}

// NewConfirm creates a Confirm dialog.
func NewConfirm(prompt, id string) Confirm {
	return Confirm{prompt: prompt, id: id, visible: true}
}

// Update handles keystrokes.
func (c Confirm) Update(msg tea.Msg) (Confirm, tea.Cmd) {
	if !c.visible {
		return c, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("y", "Y", "enter"))):
			c.visible = false
			return c, func() tea.Msg { return ConfirmedMsg{ID: c.id} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "N", "esc"))):
			c.visible = false
			return c, func() tea.Msg { return CancelledMsg{} }
		}
	}
	return c, nil
}

// View renders the dialog as an overlay string.
func (c Confirm) View() string {
	if !c.visible {
		return ""
	}
	inner := c.prompt + "\n\n  [y] Yes  [n] No"
	return confirmStyle.Render(inner)
}

// Visible returns whether the dialog is currently shown.
func (c Confirm) Visible() bool { return c.visible }
