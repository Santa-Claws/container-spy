package types

import "time"

// ContainerInfo is a flattened, safe-to-copy snapshot of a container's state.
type ContainerInfo struct {
	ID        string
	Name      string
	Image     string
	Status    string // human-readable: "Up 2 hours", "Exited (0) 3 days ago"
	State     string // raw state: "running", "exited", "paused", "restarting", "dead"
	CreatedAt time.Time
	ServerID  string
}
