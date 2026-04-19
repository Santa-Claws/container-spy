package pages

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmac/container-spy/internal/config"
	dockerpkg "github.com/tmac/container-spy/internal/docker"
	"github.com/tmac/container-spy/internal/store"
	"github.com/tmac/container-spy/internal/tui"
	"github.com/tmac/container-spy/internal/types"
)

// row is one line in the dashboard table.
type row struct {
	serverName string
	group      string
	name       string
	image      string
	status     string
	state      string
	uptime     string
	isHeader   bool   // true for group-header pseudo-rows
	headerText string // only set for header rows
}

// Dashboard is the main container overview page.
type Dashboard struct {
	width, height int
	cfg           *config.Config
	store         *store.Store
	serverNames   map[string]string // serverID → name
	rows          []row
	cursor        int
	keys          tui.KeyMap
	scrollOffset  int
}

// NewDashboard creates a Dashboard model.
func NewDashboard(cfg *config.Config, st *store.Store, keys tui.KeyMap) Dashboard {
	d := Dashboard{
		cfg:   cfg,
		store: st,
		keys:  keys,
	}
	d.rebuild()
	return d
}

// SetSize updates the terminal dimensions.
func (d *Dashboard) SetSize(w, h int) {
	d.width = w
	d.height = h
}

// Refresh signals new store data is available.
func (d *Dashboard) Refresh() {
	d.rebuild()
}

func (d *Dashboard) rebuild() {
	// Build server name lookup.
	d.serverNames = make(map[string]string, len(d.cfg.Servers))
	for _, s := range d.cfg.Servers {
		d.serverNames[s.ID] = s.Name
	}

	// Build assignment lookup: serverID+name → groupName.
	type key2 struct{ serverID, name string }
	groupOf := make(map[key2]string)
	for _, g := range d.cfg.Groups {
		for _, a := range g.Assignments {
			groupOf[key2{a.ServerID, a.ContainerName}] = g.Name
		}
	}

	snapshot := d.store.Snapshot()

	// Collect all containers.
	type entry struct {
		c     types.ContainerInfo
		group string
	}
	var entries []entry
	for _, containers := range snapshot {
		for _, c := range containers {
			g := groupOf[key2{c.ServerID, c.Name}]
			if g == "" {
				g = "Ungrouped"
			}
			entries = append(entries, entry{c, g})
		}
	}

	// Sort by group, then server name, then container name.
	sort.Slice(entries, func(i, j int) bool {
		gi, gj := entries[i].group, entries[j].group
		// Ungrouped always last.
		if gi != gj {
			if gi == "Ungrouped" {
				return false
			}
			if gj == "Ungrouped" {
				return true
			}
			return gi < gj
		}
		si := d.serverNames[entries[i].c.ServerID]
		sj := d.serverNames[entries[j].c.ServerID]
		if si != sj {
			return si < sj
		}
		return entries[i].c.Name < entries[j].c.Name
	})

	// Build rows with group headers.
	d.rows = nil
	lastGroup := ""
	for _, e := range entries {
		if e.group != lastGroup {
			d.rows = append(d.rows, row{
				isHeader:   true,
				headerText: e.group,
			})
			lastGroup = e.group
		}
		d.rows = append(d.rows, row{
			serverName: d.serverNames[e.c.ServerID],
			group:      e.group,
			name:       e.c.Name,
			image:      e.c.Image,
			status:     e.c.Status,
			state:      e.c.State,
			uptime:     formatUptime(e.c.CreatedAt),
		})
	}

	// Clamp cursor.
	if d.cursor >= len(d.rows) && len(d.rows) > 0 {
		d.cursor = len(d.rows) - 1
	}
}

func formatUptime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// Init implements tea.Model.
func (d Dashboard) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (d Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.Up):
			if d.cursor > 0 {
				d.cursor--
				// Skip headers.
				if d.rows[d.cursor].isHeader && d.cursor > 0 {
					d.cursor--
				}
			}
		case key.Matches(msg, d.keys.Down):
			if d.cursor < len(d.rows)-1 {
				d.cursor++
				if d.rows[d.cursor].isHeader && d.cursor < len(d.rows)-1 {
					d.cursor++
				}
			}
		case key.Matches(msg, d.keys.Servers):
			return d, func() tea.Msg { return tui.NavigateMsg{Page: tui.PageServers} }
		case key.Matches(msg, d.keys.Groups):
			return d, func() tea.Msg { return tui.NavigateMsg{Page: tui.PageGroups} }
		case key.Matches(msg, d.keys.Quit):
			return d, tea.Quit
		}
	case dockerpkg.RefreshMsg:
		d.rebuild()
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
	}
	return d, nil
}

// View implements tea.Model.
func (d Dashboard) View() string {
	if len(d.rows) == 0 {
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
			"\n  No containers found.\n\n  Press [s] to add a server, or wait for the first poll.\n\n  Hint: use [g] to manage groups.",
		)
		return hint
	}

	// Column widths.
	colName := 28
	colServer := 16
	colStatus := 16
	colImage := d.width - colName - colServer - colStatus - 6
	if colImage < 10 {
		colImage = 10
	}

	// Header row.
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
		colName, "CONTAINER",
		colServer, "SERVER",
		colStatus, "STATUS",
		"IMAGE",
	)

	tableHeight := d.height - 4 // title + header + statusbar + margin
	if tableHeight < 1 {
		tableHeight = 1
	}

	// Adjust scroll.
	if d.cursor-d.scrollOffset >= tableHeight {
		d.scrollOffset = d.cursor - tableHeight + 1
	}
	if d.cursor < d.scrollOffset {
		d.scrollOffset = d.cursor
	}

	var lines []string
	lines = append(lines, headerStyle.Render(header))

	for i := d.scrollOffset; i < len(d.rows) && i < d.scrollOffset+tableHeight; i++ {
		r := d.rows[i]
		if r.isHeader {
			lines = append(lines, "")
			lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).
				Render("  ▸ "+r.headerText))
			continue
		}

		stateStyle := tui.StateStyle(r.state)
		statusCell := stateStyle.Render(truncate(r.status, colStatus))
		nameCell := truncate(r.name, colName)
		serverCell := truncate(r.serverName, colServer)
		imageCell := truncate(r.image, colImage)

		line := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
			colName, nameCell,
			colServer, serverCell,
			colStatus, statusCell,
			imageCell,
		)

		if i == d.cursor {
			line = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("15")).Render(line)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
