package docker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	dcontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/tmac/container-spy/internal/config"
	"github.com/tmac/container-spy/internal/store"
	"github.com/tmac/container-spy/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

// RefreshMsg is sent to the Bubble Tea program after each poll cycle.
type RefreshMsg struct{}

// Poller polls a single server's Docker daemon on a fixed interval.
type Poller struct {
	server   config.Server
	store    *store.Store
	interval time.Duration
	stop     chan struct{}
	mu       sync.Mutex
	prog     *tea.Program
}

// Manager manages a set of pollers, one per server.
type Manager struct {
	mu      sync.Mutex
	pollers map[string]*Poller
	store   *store.Store
	prog    *tea.Program
}

// NewManager creates a new poller manager.
func NewManager(st *store.Store) *Manager {
	return &Manager{
		pollers: make(map[string]*Poller),
		store:   st,
	}
}

// SetProgram sets the Bubble Tea program reference (call after tea.NewProgram).
func (m *Manager) SetProgram(p *tea.Program) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prog = p
	for _, p2 := range m.pollers {
		p2.setProgram(p)
	}
}

// Sync starts/stops pollers to match the provided server list.
func (m *Manager) Sync(servers []config.Server) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wanted := make(map[string]config.Server, len(servers))
	for _, s := range servers {
		wanted[s.ID] = s
	}

	// Stop pollers for removed servers.
	for id, p := range m.pollers {
		if _, ok := wanted[id]; !ok {
			p.Stop()
			delete(m.pollers, id)
		}
	}

	// Start pollers for new servers.
	for id, srv := range wanted {
		if _, exists := m.pollers[id]; !exists {
			p := &Poller{
				server:   srv,
				store:    m.store,
				interval: 30 * time.Second,
				stop:     make(chan struct{}),
				prog:     m.prog,
			}
			m.pollers[id] = p
			go p.run()
		}
	}
}

// StopAll stops all pollers.
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.pollers {
		p.Stop()
	}
	m.pollers = make(map[string]*Poller)
}

func (p *Poller) setProgram(prog *tea.Program) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prog = prog
}

// Stop signals the poller to exit.
func (p *Poller) Stop() {
	select {
	case <-p.stop:
	default:
		close(p.stop)
	}
}

func (p *Poller) run() {
	// Poll immediately on start.
	p.poll()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	backoff := time.Second
	maxBackoff := 5 * time.Minute

	for {
		select {
		case <-p.stop:
			return
		case <-ticker.C:
			if err := p.poll(); err != nil {
				slog.Warn("poll failed, backing off",
					"server", p.server.Name, "err", err, "backoff", backoff)
				time.Sleep(backoff)
				backoff = min(backoff*2, maxBackoff)
			} else {
				backoff = time.Second
			}
		}
	}
}

func (p *Poller) poll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cl, err := newDockerClient(p.server)
	if err != nil {
		p.store.SetError(p.server.ID, err)
		p.notify()
		return err
	}
	defer cl.Close()

	list, err := cl.ContainerList(ctx, dcontainer.ListOptions{All: true})
	if err != nil {
		p.store.SetError(p.server.ID, err)
		p.notify()
		return err
	}

	containers := make([]types.ContainerInfo, 0, len(list))
	for _, c := range list {
		containers = append(containers, FromSDK(c, p.server.ID))
	}
	p.store.Set(p.server.ID, containers)
	p.notify()
	return nil
}

func (p *Poller) notify() {
	p.mu.Lock()
	prog := p.prog
	p.mu.Unlock()
	if prog != nil {
		prog.Send(RefreshMsg{})
	}
}

// newDockerClient creates a client; extracted so it can be replaced in tests.
var newDockerClient = func(srv config.Server) (*dockerclient.Client, error) {
	return NewClient(srv)
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
