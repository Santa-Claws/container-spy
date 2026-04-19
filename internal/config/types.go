package config

// Config is the root configuration structure persisted to YAML.
type Config struct {
	Servers []Server `yaml:"servers"`
	Groups  []Group  `yaml:"groups"`
}

// Server represents a remote machine accessible via SSH.
type Server struct {
	ID     string `yaml:"id"`      // UUID
	Name   string `yaml:"name"`    // friendly name, e.g. "web-01"
	Host   string `yaml:"host"`    // IP or hostname
	User   string `yaml:"user"`    // SSH user
	SSHKey string `yaml:"ssh_key"` // absolute path inside container, e.g. /keys/id_rsa
}

// Group is a named collection of container assignments.
type Group struct {
	ID          string       `yaml:"id"`
	Name        string       `yaml:"name"`
	Assignments []Assignment `yaml:"assignments"`
}

// Assignment links a container name on a specific server to a group.
type Assignment struct {
	ServerID      string `yaml:"server_id"`
	ContainerName string `yaml:"container_name"` // matched by name (stable across restarts)
}
