package docker

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/tmac/container-spy/internal/types"
)

// ContainerInfo is an alias for the shared type.
type ContainerInfo = types.ContainerInfo

// FromSDK converts a Docker SDK container summary to ContainerInfo.
func FromSDK(c container.Summary, serverID string) types.ContainerInfo {
	name := ""
	if len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}
	return types.ContainerInfo{
		ID:        shortID(c.ID),
		Name:      name,
		Image:     c.Image,
		Status:    c.Status,
		State:     c.State,
		CreatedAt: time.Unix(c.Created, 0),
		ServerID:  serverID,
	}
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
