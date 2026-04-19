package docker

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
	"github.com/tmac/container-spy/internal/config"
)

// NewClient creates a Docker client that connects to the server via SSH.
// SSH transport is provided by docker/cli's connhelper, which shells out to
// the system ssh binary and honours ~/.ssh/config.
// GenerateSSHConfig must be called before this to write the correct key/user
// entries for each host.
func NewClient(srv config.Server) (*client.Client, error) {
	host := fmt.Sprintf("ssh://%s@%s", srv.User, srv.Host)
	helper, err := connhelper.GetConnectionHelper(host)
	if err != nil {
		return nil, fmt.Errorf("ssh connection helper: %w", err)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: helper.Dialer,
		},
	}
	return client.NewClientWithOpts(
		client.WithHTTPClient(httpClient),
		client.WithHost(helper.Host),
		client.WithAPIVersionNegotiation(),
	)
}

// GenerateSSHConfig writes (or overwrites) a ~/.ssh/config block for every
// server in cfg. Call this at startup and whenever the server list changes.
func GenerateSSHConfig(servers []config.Server) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# container-spy managed — do not edit\n\n")
	for _, srv := range servers {
		key := srv.SSHKey
		if key == "" {
			key = filepath.Join(home, ".ssh", "id_rsa")
		}
		// Warn if key permissions are too open.
		if info, err := os.Stat(key); err == nil {
			if info.Mode().Perm() > 0600 {
				slog.Warn("SSH key permissions may be too open; SSH may reject it",
					"path", key, "mode", fmt.Sprintf("%o", info.Mode().Perm()))
			}
		}
		fmt.Fprintf(&b, "Host %s\n", srv.Host)
		fmt.Fprintf(&b, "    User %s\n", srv.User)
		fmt.Fprintf(&b, "    IdentityFile %s\n", key)
		fmt.Fprintf(&b, "    StrictHostKeyChecking accept-new\n")
		fmt.Fprintf(&b, "    ConnectTimeout 10\n")
		fmt.Fprintf(&b, "    ServerAliveInterval 30\n\n")
	}

	cfgPath := filepath.Join(sshDir, "config")
	return os.WriteFile(cfgPath, []byte(b.String()), 0600)
}
