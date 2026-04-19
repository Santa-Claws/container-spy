# container-spy

Monitor Docker containers across multiple remote machines from a single terminal or browser.

Connects via Docker-over-SSH — **no agents, no daemon config, no changes needed on remote machines**. Just SSH access.

## Features

- **TUI** (terminal) or **WebUI** (browser) — both built in
- Monitor containers on any number of remote machines
- Name your servers and organize containers into named **groups** (e.g. "nginx", "databases")
- Live status with colour coding: running / exited / paused
- All configuration done through the TUI — no manual YAML editing needed
- Ships as a single Docker container

## How it works

container-spy connects to each remote Docker daemon over SSH using your existing keys. It generates the necessary SSH config automatically at startup. Every 30 seconds it polls each server for the current container list and updates the display.

```
your machine
└── container-spy (docker container)
    ├── SSH → web-01 → docker daemon → containers
    ├── SSH → db-01  → docker daemon → containers
    └── SSH → dev-01 → docker daemon → containers
```

## Requirements

- Docker + Docker Compose on the machine running container-spy
- SSH key access to each remote machine (the Docker daemon must be accessible — typically via the Unix socket at `/var/run/docker.sock`, which is the default)
- The remote user must have permission to access Docker (be in the `docker` group or run as root)

## Setup

### 1. Clone the repo

```sh
git clone https://github.com/Santa-Claws/container-spy.git
cd container-spy
```

> **Important:** all commands below must be run from the repo root (`container-spy/`). If you `cd` into a subdirectory and run `docker compose`, it will fail with "no configuration file provided".

### 2. Add your SSH keys

From the repo root, create a `keys/` directory and copy in the private key(s) you use to access your servers:

```sh
mkdir keys
cp ~/.ssh/id_ed25519 keys/id_ed25519   # or id_rsa, etc.
chmod 600 keys/*
```

You can add multiple keys; you'll specify which key to use per server in the TUI.

### 3. Create the config directory

```sh
mkdir config
```

This is where container-spy stores its configuration (servers, groups). It starts empty — everything is configured through the TUI.

### 4. Build the image

```sh
docker compose build
# or: make build-image
```

### 5. Run in TUI mode

```sh
docker compose run --rm container-spy
# or: make run-tui
```

You must use `docker compose run` (not `up`) for TUI mode so it attaches to your terminal correctly.

### 6. Add your first server

Inside the TUI, press **`s`** to open the Servers page, then **`a`** to add a server:

| Field | Example |
|---|---|
| Name | `web-01` |
| Host / IP | `192.168.1.10` |
| SSH User | `ubuntu` |
| SSH Key Path | `/keys/id_rsa` |

Press **Tab** to advance between fields, **Enter** on the last field to save.

container-spy will immediately begin polling that server. Press **`d`** to return to the dashboard.

### 7. Organize into groups (optional)

Press **`g`** to open the Groups page:
- **`a`** to create a new group (e.g. "nginx", "databases")
- **Tab** or **→** to switch to the container list on the right
- **Space** to toggle a container's membership in the selected group

Group assignments are saved instantly.

## Running the WebUI

To expose a browser-accessible dashboard instead of a TUI:

```sh
CONTAINER_SPY_MODE=web docker compose up -d
# or: make run-web
```

Then open [http://localhost:8080](http://localhost:8080). The page polls for updates every 5 seconds.

If port 8080 is already in use, override it:

```sh
CONTAINER_SPY_MODE=web CONTAINER_SPY_ADDR=:9090 docker compose up -d
```

The web dashboard shows the same grouped view as the TUI. Configuration (adding servers, managing groups) is TUI-only.

## TUI keybindings

| Key | Action |
|---|---|
| `d` | Dashboard |
| `s` | Servers page |
| `g` | Groups page |
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `←` / `→` | Switch pane (groups page) |
| `a` | Add |
| `e` | Edit |
| `D` | Delete |
| `Space` | Toggle group assignment |
| `Tab` | Next field (forms) / next pane |
| `Enter` | Confirm |
| `Esc` | Cancel / back |
| `q` | Quit |

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `CONTAINER_SPY_MODE` | `tui` | `tui` or `web` |
| `CONTAINER_SPY_CONFIG` | `/config/config.yaml` | Path to config file |
| `CONTAINER_SPY_ADDR` | `:8080` | Listen address (web mode only) |

## Volume mounts

| Path | Description |
|---|---|
| `/keys` | SSH private keys (mount read-only) |
| `/config` | Persistent config directory (must be writable) |

## Logs

In TUI mode, logs are written to `/tmp/container-spy.log` inside the container (so they don't corrupt the terminal). In web mode, logs go to stdout.

## Building locally (without Docker)

```sh
go build -o container-spy ./cmd/container-spy

# TUI
./container-spy

# WebUI
CONTAINER_SPY_MODE=web CONTAINER_SPY_ADDR=:8080 ./container-spy
```

Requires Go 1.21+ and `openssh-client` installed on the host.

## Config file format

The config is stored as YAML and managed by the TUI. You can inspect it at `config/config.yaml`:

```yaml
servers:
  - id: "abc-123"
    name: "web-01"
    host: "192.168.1.10"
    user: "ubuntu"
    ssh_key: "/keys/id_rsa"

groups:
  - id: "def-456"
    name: "nginx"
    assignments:
      - server_id: "abc-123"
        container_name: "nginx-proxy"
```

Container assignments are matched by name (stable across restarts when using Docker Compose) rather than by ID.
