package components

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormSubmittedMsg carries the submitted field values.
type FormSubmittedMsg struct{ Values []string }

// FormCancelledMsg signals the user pressed Escape.
type FormCancelledMsg struct{}

var (
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Width(14)
	activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
)

// Field is a single labeled text input.
type Field struct {
	Label    string
	input    textinput.Model
}

// Form renders a stack of labeled text inputs.
type Form struct {
	fields  []Field
	focused int
}

// NewForm creates a form with the given labeled fields and optional initial values.
func NewForm(labels []string, values []string) Form {
	fields := make([]Field, len(labels))
	for i, lbl := range labels {
		ti := textinput.New()
		ti.Placeholder = lbl
		ti.CharLimit = 256
		if i < len(values) {
			ti.SetValue(values[i])
		}
		if i == 0 {
			ti.Focus()
		}
		fields[i] = Field{Label: lbl, input: ti}
	}
	return Form{fields: fields}
}

// Update handles keyboard events for the form.
func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return f, func() tea.Msg { return FormCancelledMsg{} }

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "down"))):
			f.fields[f.focused].input.Blur()
			f.focused = (f.focused + 1) % len(f.fields)
			f.fields[f.focused].input.Focus()
			return f, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "up"))):
			f.fields[f.focused].input.Blur()
			f.focused = (f.focused - 1 + len(f.fields)) % len(f.fields)
			f.fields[f.focused].input.Focus()
			return f, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if f.focused == len(f.fields)-1 {
				vals := make([]string, len(f.fields))
				for i, fd := range f.fields {
					vals[i] = fd.input.Value()
				}
				return f, func() tea.Msg { return FormSubmittedMsg{Values: vals} }
			}
			// advance to next field
			f.fields[f.focused].input.Blur()
			f.focused++
			f.fields[f.focused].input.Focus()
			return f, nil
		}
	}

	var cmd tea.Cmd
	f.fields[f.focused].input, cmd = f.fields[f.focused].input.Update(msg)
	return f, cmd
}

// View renders the form.
func (f Form) View() string {
	var out string
	for i, fd := range f.fields {
		label := labelStyle.Render(fd.Label + ":")
		inp := fd.input.View()
		if i == f.focused {
			inp = activeStyle.Render(inp)
		}
		out += label + " " + inp + "\n"
	}
	out += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("tab: next field  enter: submit  esc: cancel")
	return out
}

// Values returns the current input values.
func (f Form) Values() []string {
	vals := make([]string, len(f.fields))
	for i, fd := range f.fields {
		vals[i] = fd.input.Value()
	}
	return vals
}
