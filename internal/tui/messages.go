package tui

// PageID identifies a TUI page.
type PageID string

const (
	PageDashboard PageID = "dashboard"
	PageServers   PageID = "servers"
	PageGroups    PageID = "groups"
)

// NavigateMsg switches the active page.
type NavigateMsg struct{ Page PageID }

// SaveConfigMsg signals that the config should be persisted.
type SaveConfigMsg struct{}

// ServersSyncMsg signals that the poller manager should re-sync servers.
type ServersSyncMsg struct{}

// ErrorMsg carries a displayable error.
type ErrorMsg struct{ Err error }
