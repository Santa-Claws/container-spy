package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmac/container-spy/internal/config"
	dockerpkg "github.com/tmac/container-spy/internal/docker"
	"github.com/tmac/container-spy/internal/store"
	"github.com/tmac/container-spy/internal/tui/components"
)

// Page is the interface every page model must satisfy.
type Page interface {
	tea.Model
	SetSize(w, h int)
}

// Model is the root Bubble Tea model.
type Model struct {
	width, height int
	cfg           *config.Config
	cfgPath       string
	store         *store.Store
	keys          KeyMap
	currentPage   PageID
	pages         map[PageID]Page
	statusBar     components.StatusBar
	lastError     string
	// OnSync is called when a ServersSyncMsg is received (e.g. to restart pollers).
	OnSync func()
}

// PageFactory is a function type to break import cycles.
type PageFactory func(cfg *config.Config, cfgPath string, st *store.Store, keys KeyMap) map[PageID]Page

// New creates the root TUI model with pre-built pages.
func New(cfg *config.Config, cfgPath string, st *store.Store, pages map[PageID]Page) Model {
	keys := DefaultKeyMap()
	return Model{
		cfg:         cfg,
		cfgPath:     cfgPath,
		store:       st,
		keys:        keys,
		currentPage: PageDashboard,
		pages:       pages,
		statusBar:   components.NewStatusBar("tui"),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward to all pages so they always have current dimensions.
		for _, p := range m.pages {
			p.SetSize(msg.Width, msg.Height-3) // -3 for title + statusbar
		}

	case NavigateMsg:
		m.currentPage = msg.Page

	case dockerpkg.RefreshMsg:
		// Update status bar.
		errs := m.store.Errors()
		errCount := 0
		var latestRefresh time.Time
		for _, t := range m.store.LastRefresh() {
			if t.After(latestRefresh) {
				latestRefresh = t
			}
		}
		for _, e := range errs {
			if e != nil {
				errCount++
			}
		}
		hint := "[s] servers  [g] groups  [d] dashboard  [q] quit"
		m.statusBar.Update(latestRefresh, errCount, hint)

	case ErrorMsg:
		m.lastError = msg.Err.Error()

	case ServersSyncMsg:
		if m.OnSync != nil {
			go m.OnSync()
		}
	}

	// Delegate to current page.
	cur := m.pages[m.currentPage]
	updated, cmd := cur.Update(msg)
	m.pages[m.currentPage] = updated.(Page)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).
		Background(lipgloss.Color("236")).Width(m.width).Padding(0, 1)
	title := titleStyle.Render("container-spy")

	content := ""
	if p, ok := m.pages[m.currentPage]; ok {
		content = p.View()
	}

	errLine := ""
	if m.lastError != "" {
		errLine = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).
			Render(fmt.Sprintf("  Error: %s", m.lastError))
	}

	m.statusBar.SetWidth(m.width)
	bar := m.statusBar.View()

	// Assemble: title + content area + status bar.
	contentHeight := m.height - 2 // title + bar
	if contentHeight < 0 {
		contentHeight = 0
	}
	_ = contentHeight

	return title + "\n" + content + errLine + "\n" + bar
}
