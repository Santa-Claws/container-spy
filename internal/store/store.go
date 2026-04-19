package store

import (
	"sync"
	"time"

	"github.com/tmac/container-spy/internal/types"
)

// Store is a thread-safe snapshot store keyed by server ID.
type Store struct {
	mu          sync.RWMutex
	data        map[string][]types.ContainerInfo
	errs        map[string]error
	lastRefresh map[string]time.Time
}

// New creates an empty Store.
func New() *Store {
	return &Store{
		data:        make(map[string][]types.ContainerInfo),
		errs:        make(map[string]error),
		lastRefresh: make(map[string]time.Time),
	}
}

// Set updates the container list for serverID.
func (s *Store) Set(serverID string, containers []types.ContainerInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[serverID] = containers
	s.errs[serverID] = nil
	s.lastRefresh[serverID] = time.Now()
}

// SetError records a poll error for serverID.
func (s *Store) SetError(serverID string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errs[serverID] = err
	s.lastRefresh[serverID] = time.Now()
}

// Snapshot returns a deep copy of all container data. Callers can safely
// read the result without holding any lock.
func (s *Store) Snapshot() map[string][]types.ContainerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]types.ContainerInfo, len(s.data))
	for k, v := range s.data {
		cp := make([]types.ContainerInfo, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

// Errors returns a copy of the current error map.
func (s *Store) Errors() map[string]error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]error, len(s.errs))
	for k, v := range s.errs {
		out[k] = v
	}
	return out
}

// LastRefresh returns a copy of the last-refresh time map.
func (s *Store) LastRefresh() map[string]time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]time.Time, len(s.lastRefresh))
	for k, v := range s.lastRefresh {
		out[k] = v
	}
	return out
}
