package pages

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/tmac/container-spy/internal/config"
	"github.com/tmac/container-spy/internal/tui"
	"github.com/tmac/container-spy/internal/tui/components"
)

type serverMode int

const (
	serverList serverMode = iota
	serverAdd
	serverEdit
	serverConfirmDelete
)

// Servers is the server management page.
type Servers struct {
	width, height int
	cfg           *config.Config
	cfgPath       string
	keys          tui.KeyMap
	cursor        int
	mode          serverMode
	form          components.Form
	confirm       components.Confirm
	editIdx       int
}

// NewServers creates a Servers page.
func NewServers(cfg *config.Config, cfgPath string, keys tui.KeyMap) Servers {
	return Servers{
		cfg:     cfg,
		cfgPath: cfgPath,
		keys:    keys,
	}
}

// SetSize updates dimensions.
func (s *Servers) SetSize(w, h int) { s.width = w; s.height = h }

func (s Servers) Init() tea.Cmd { return nil }

func (s Servers) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle form/confirm message results first (they arrive as messages in the next Update cycle).
	switch msg := msg.(type) {
	case components.FormSubmittedMsg:
		s.applyForm(msg.Values)
		s.mode = serverList
		return s, func() tea.Msg { return tui.ServersSyncMsg{} }

	case components.FormCancelledMsg:
		s.mode = serverList
		return s, nil

	case components.ConfirmedMsg:
		s.deleteServer(msg.ID)
		s.mode = serverList
		return s, func() tea.Msg { return tui.ServersSyncMsg{} }

	case components.CancelledMsg:
		s.mode = serverList
		return s, nil
	}

	switch s.mode {
	case serverConfirmDelete:
		var cmd tea.Cmd
		s.confirm, cmd = s.confirm.Update(msg)
		if !s.confirm.Visible() {
			s.mode = serverList
		}
		return s, cmd

	case serverAdd, serverEdit:
		var cmd tea.Cmd
		s.form, cmd = s.form.Update(msg)
		return s, cmd
	}

	// List mode.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keys.Up):
			if s.cursor > 0 {
				s.cursor--
			}
		case key.Matches(msg, s.keys.Down):
			if s.cursor < len(s.cfg.Servers)-1 {
				s.cursor++
			}
		case key.Matches(msg, s.keys.Add):
			s.form = components.NewForm(
				[]string{"Name", "Host / IP", "SSH User", "SSH Key Path"},
				[]string{"", "", "root", "/keys/id_rsa"},
			)
			s.mode = serverAdd
		case key.Matches(msg, s.keys.Edit):
			if len(s.cfg.Servers) > 0 {
				srv := s.cfg.Servers[s.cursor]
				s.form = components.NewForm(
					[]string{"Name", "Host / IP", "SSH User", "SSH Key Path"},
					[]string{srv.Name, srv.Host, srv.User, srv.SSHKey},
				)
				s.editIdx = s.cursor
				s.mode = serverEdit
			}
		case key.Matches(msg, s.keys.Delete):
			if len(s.cfg.Servers) > 0 {
				srv := s.cfg.Servers[s.cursor]
				s.confirm = components.NewConfirm(
					fmt.Sprintf("Delete server %q?", srv.Name), srv.ID,
				)
				s.mode = serverConfirmDelete
			}
		case key.Matches(msg, s.keys.Dashboard):
			return s, func() tea.Msg { return tui.NavigateMsg{Page: tui.PageDashboard} }
		case key.Matches(msg, s.keys.Groups):
			return s, func() tea.Msg { return tui.NavigateMsg{Page: tui.PageGroups} }
		case key.Matches(msg, s.keys.Quit):
			return s, tea.Quit
		}
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}
	return s, nil
}

func (s *Servers) applyForm(vals []string) {
	if len(vals) < 4 {
		return
	}
	if s.mode == serverAdd {
		s.cfg.Servers = append(s.cfg.Servers, config.Server{
			ID:     uuid.New().String(),
			Name:   vals[0],
			Host:   vals[1],
			User:   vals[2],
			SSHKey: vals[3],
		})
	} else {
		if s.editIdx < len(s.cfg.Servers) {
			s.cfg.Servers[s.editIdx].Name = vals[0]
			s.cfg.Servers[s.editIdx].Host = vals[1]
			s.cfg.Servers[s.editIdx].User = vals[2]
			s.cfg.Servers[s.editIdx].SSHKey = vals[3]
		}
	}
	config.Save(s.cfg, s.cfgPath) //nolint:errcheck
}

func (s *Servers) deleteServer(id string) {
	for i, srv := range s.cfg.Servers {
		if srv.ID == id {
			s.cfg.Servers = append(s.cfg.Servers[:i], s.cfg.Servers[i+1:]...)
			if s.cursor >= len(s.cfg.Servers) && s.cursor > 0 {
				s.cursor--
			}
			config.Save(s.cfg, s.cfgPath) //nolint:errcheck
			return
		}
	}
}

func (s Servers) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("  Servers")

	if s.mode == serverAdd || s.mode == serverEdit {
		action := "Add Server"
		if s.mode == serverEdit {
			action = "Edit Server"
		}
		header := lipgloss.NewStyle().Bold(true).Render("  " + action)
		return title + "\n\n" + header + "\n\n" + s.form.View()
	}

	if s.mode == serverConfirmDelete {
		base := s.renderList()
		overlay := lipgloss.Place(s.width, s.height/2,
			lipgloss.Center, lipgloss.Center, s.confirm.View())
		return title + "\n\n" + base + "\n" + overlay
	}

	return title + "\n\n" + s.renderList() + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
			"  [a] add  [e] edit  [D] delete  [d] dashboard  [g] groups  [q] quit",
		)
}

func (s Servers) renderList() string {
	if len(s.cfg.Servers) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
			"  No servers configured. Press [a] to add one.",
		)
	}

	var lines []string
	for i, srv := range s.cfg.Servers {
		line := fmt.Sprintf("  %-20s  %-20s  %s  %s",
			srv.Name, srv.Host, srv.User, srv.SSHKey)
		if i == s.cursor {
			line = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("15")).Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
