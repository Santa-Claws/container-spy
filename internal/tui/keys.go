package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds the global keybindings.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	Enter    key.Binding
	Escape   key.Binding
	Quit     key.Binding
	Servers  key.Binding
	Groups   key.Binding
	Dashboard key.Binding
	Add      key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Save     key.Binding
	Space    key.Binding
	Help     key.Binding
}

// DefaultKeyMap returns the standard keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:      key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:     key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
		Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Escape:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel/back")),
		Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Servers:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "servers")),
		Groups:    key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "groups")),
		Dashboard: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "dashboard")),
		Add:       key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		Edit:      key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Delete:    key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "delete")),
		Save:      key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
		Space:     key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}
