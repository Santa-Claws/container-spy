package pages

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/tmac/container-spy/internal/config"
	dockerpkg "github.com/tmac/container-spy/internal/docker"
	"github.com/tmac/container-spy/internal/store"
	"github.com/tmac/container-spy/internal/tui"
	"github.com/tmac/container-spy/internal/tui/components"
	"github.com/tmac/container-spy/internal/types"
)

type groupPane int

const (
	groupPaneLeft  groupPane = iota // groups list
	groupPaneRight                  // container picker
)

type groupMode int

const (
	groupModeView   groupMode = iota
	groupModeAdd              // entering new group name
	groupModeDelete           // confirm delete
)

// Groups is the group management page.
type Groups struct {
	width, height int
	cfg           *config.Config
	cfgPath       string
	store         *store.Store
	keys          tui.KeyMap

	activePane groupPane
	groupIdx   int // selected group in left pane
	contIdx    int // selected container in right pane

	// flat container list from store (for right pane)
	allContainers []types.ContainerInfo
	serverNames   map[string]string

	mode    groupMode
	form    components.Form
	confirm components.Confirm
}

// NewGroups creates a Groups page.
func NewGroups(cfg *config.Config, cfgPath string, st *store.Store, keys tui.KeyMap) Groups {
	g := Groups{
		cfg:     cfg,
		cfgPath: cfgPath,
		store:   st,
		keys:    keys,
	}
	g.refreshContainers()
	return g
}

// SetSize updates dimensions.
func (g *Groups) SetSize(w, h int) { g.width = w; g.height = h }

func (g *Groups) refreshContainers() {
	g.serverNames = make(map[string]string, len(g.cfg.Servers))
	for _, s := range g.cfg.Servers {
		g.serverNames[s.ID] = s.Name
	}

	snapshot := g.store.Snapshot()
	g.allContainers = nil
	for _, cs := range snapshot {
		g.allContainers = append(g.allContainers, cs...)
	}
	sort.Slice(g.allContainers, func(i, j int) bool {
		si := g.serverNames[g.allContainers[i].ServerID]
		sj := g.serverNames[g.allContainers[j].ServerID]
		if si != sj {
			return si < sj
		}
		return g.allContainers[i].Name < g.allContainers[j].Name
	})
}

func (g Groups) Init() tea.Cmd { return nil }

func (g Groups) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle form/confirm message results (they arrive in the next Update cycle).
	switch msg := msg.(type) {
	case components.FormSubmittedMsg:
		if len(msg.Values) > 0 && msg.Values[0] != "" {
			g.cfg.Groups = append(g.cfg.Groups, config.Group{
				ID:   uuid.New().String(),
				Name: msg.Values[0],
			})
			g.groupIdx = len(g.cfg.Groups) - 1
			config.Save(g.cfg, g.cfgPath) //nolint:errcheck
		}
		g.mode = groupModeView
		return g, nil

	case components.FormCancelledMsg:
		g.mode = groupModeView
		return g, nil

	case components.ConfirmedMsg:
		g.deleteGroup(msg.ID)
		g.mode = groupModeView
		return g, nil

	case components.CancelledMsg:
		g.mode = groupModeView
		return g, nil
	}

	switch g.mode {
	case groupModeAdd:
		var cmd tea.Cmd
		g.form, cmd = g.form.Update(msg)
		return g, cmd

	case groupModeDelete:
		var cmd tea.Cmd
		g.confirm, cmd = g.confirm.Update(msg)
		if !g.confirm.Visible() {
			g.mode = groupModeView
		}
		return g, cmd
	}

	// View mode.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, g.keys.Left):
			g.activePane = groupPaneLeft
		case key.Matches(msg, g.keys.Right), key.Matches(msg, g.keys.Tab):
			g.activePane = groupPaneRight

		case key.Matches(msg, g.keys.Up):
			if g.activePane == groupPaneLeft {
				if g.groupIdx > 0 {
					g.groupIdx--
				}
			} else {
				if g.contIdx > 0 {
					g.contIdx--
				}
			}

		case key.Matches(msg, g.keys.Down):
			if g.activePane == groupPaneLeft {
				if g.groupIdx < len(g.cfg.Groups)-1 {
					g.groupIdx++
				}
			} else {
				if g.contIdx < len(g.allContainers)-1 {
					g.contIdx++
				}
			}

		case key.Matches(msg, g.keys.Add):
			if g.activePane == groupPaneLeft {
				g.form = components.NewForm([]string{"Group Name"}, nil)
				g.mode = groupModeAdd
			}

		case key.Matches(msg, g.keys.Delete):
			if g.activePane == groupPaneLeft && len(g.cfg.Groups) > 0 {
				grp := g.cfg.Groups[g.groupIdx]
				g.confirm = components.NewConfirm(
					fmt.Sprintf("Delete group %q?", grp.Name), grp.ID,
				)
				g.mode = groupModeDelete
			}

		case key.Matches(msg, g.keys.Space):
			if g.activePane == groupPaneRight && len(g.cfg.Groups) > 0 && len(g.allContainers) > 0 {
				g.toggleAssignment()
			}

		case key.Matches(msg, g.keys.Dashboard):
			return g, func() tea.Msg { return tui.NavigateMsg{Page: tui.PageDashboard} }
		case key.Matches(msg, g.keys.Servers):
			return g, func() tea.Msg { return tui.NavigateMsg{Page: tui.PageServers} }
		case key.Matches(msg, g.keys.Quit):
			return g, tea.Quit
		}

	case dockerpkg.RefreshMsg:
		g.refreshContainers()

	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height
	}
	return g, nil
}

func (g *Groups) toggleAssignment() {
	if g.groupIdx >= len(g.cfg.Groups) || g.contIdx >= len(g.allContainers) {
		return
	}
	grp := &g.cfg.Groups[g.groupIdx]
	c := g.allContainers[g.contIdx]

	// Check if already assigned.
	for i, a := range grp.Assignments {
		if a.ServerID == c.ServerID && a.ContainerName == c.Name {
			// Remove.
			grp.Assignments = append(grp.Assignments[:i], grp.Assignments[i+1:]...)
			config.Save(g.cfg, g.cfgPath) //nolint:errcheck
			return
		}
	}
	// Add.
	grp.Assignments = append(grp.Assignments, config.Assignment{
		ServerID:      c.ServerID,
		ContainerName: c.Name,
	})
	config.Save(g.cfg, g.cfgPath) //nolint:errcheck
}

func (g *Groups) deleteGroup(id string) {
	for i, grp := range g.cfg.Groups {
		if grp.ID == id {
			g.cfg.Groups = append(g.cfg.Groups[:i], g.cfg.Groups[i+1:]...)
			if g.groupIdx >= len(g.cfg.Groups) && g.groupIdx > 0 {
				g.groupIdx--
			}
			config.Save(g.cfg, g.cfgPath) //nolint:errcheck
			return
		}
	}
}

func (g Groups) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("  Groups")

	if g.mode == groupModeAdd {
		return title + "\n\n" + g.form.View()
	}

	halfW := g.width/2 - 2
	if halfW < 20 {
		halfW = 20
	}

	left := g.renderGroupPane(halfW)
	right := g.renderContainerPane(halfW)

	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(halfW+2).Render(left),
		lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			PaddingLeft(1).
			Width(halfW).
			Render(right),
	)

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
		"  ←/→ switch pane  ↑/↓ navigate  [a] add group  [D] delete group  [space] toggle assignment  [d] dashboard",
	)

	if g.mode == groupModeDelete {
		overlay := lipgloss.Place(g.width, 6, lipgloss.Center, lipgloss.Center, g.confirm.View())
		return title + "\n\n" + panes + "\n" + overlay + "\n" + help
	}

	return title + "\n\n" + panes + "\n\n" + help
}

func (g Groups) renderGroupPane(w int) string {
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8")).Render("GROUPS")
	if len(g.cfg.Groups) == 0 {
		return header + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("  (none) press [a]")
	}

	var lines []string
	lines = append(lines, header)
	for i, grp := range g.cfg.Groups {
		count := len(grp.Assignments)
		line := fmt.Sprintf("%-*s  (%d)", w-8, grp.Name, count)
		if len(line) > w {
			line = line[:w]
		}
		if i == g.groupIdx && g.activePane == groupPaneLeft {
			line = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("15")).Render(line)
		} else if i == g.groupIdx {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (g Groups) renderContainerPane(w int) string {
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8")).Render("CONTAINERS  [space=toggle]")

	// Build current group's assignment set for quick lookup.
	type k2 struct{ serverID, name string }
	assigned := make(map[k2]bool)
	if g.groupIdx < len(g.cfg.Groups) {
		for _, a := range g.cfg.Groups[g.groupIdx].Assignments {
			assigned[k2{a.ServerID, a.ContainerName}] = true
		}
	}

	if len(g.allContainers) == 0 {
		return header + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("  (no containers yet)")
	}

	var lines []string
	lines = append(lines, header)
	for i, c := range g.allContainers {
		srvName := g.serverNames[c.ServerID]
		marker := " "
		if assigned[k2{c.ServerID, c.Name}] {
			marker = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓")
		}
		line := fmt.Sprintf("%s %-*s  %s", marker, w-len(srvName)-5, c.Name, srvName)
		if len(line) > w {
			line = line[:w]
		}
		if i == g.contIdx && g.activePane == groupPaneRight {
			line = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("15")).Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
