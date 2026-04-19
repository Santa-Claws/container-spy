package web

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/tmac/container-spy/internal/config"
	"github.com/tmac/container-spy/internal/store"
)

type containerResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Group      string `json:"group"`
	ServerName string `json:"server_name"`
	ServerHost string `json:"server_host"`
	Image      string `json:"image"`
	Status     string `json:"status"`
	State      string `json:"state"`
}

func buildContainerList(cfg *config.Config, st *store.Store) []containerResponse {
	// Server lookups.
	serverName := make(map[string]string, len(cfg.Servers))
	serverHost := make(map[string]string, len(cfg.Servers))
	for _, s := range cfg.Servers {
		serverName[s.ID] = s.Name
		serverHost[s.ID] = s.Host
	}

	// Assignment lookup.
	type k2 struct{ serverID, name string }
	groupOf := make(map[k2]string)
	for _, g := range cfg.Groups {
		for _, a := range g.Assignments {
			groupOf[k2{a.ServerID, a.ContainerName}] = g.Name
		}
	}

	snapshot := st.Snapshot()

	var rows []containerResponse
	for _, containers := range snapshot {
		for _, c := range containers {
			grp := groupOf[k2{c.ServerID, c.Name}]
			if grp == "" {
				grp = "Ungrouped"
			}
			rows = append(rows, containerResponse{
				ID:         c.ID,
				Name:       c.Name,
				Group:      grp,
				ServerName: serverName[c.ServerID],
				ServerHost: serverHost[c.ServerID],
				Image:      c.Image,
				Status:     c.Status,
				State:      c.State,
			})
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Group != rows[j].Group {
			if rows[i].Group == "Ungrouped" {
				return false
			}
			if rows[j].Group == "Ungrouped" {
				return true
			}
			return rows[i].Group < rows[j].Group
		}
		if rows[i].ServerName != rows[j].ServerName {
			return rows[i].ServerName < rows[j].ServerName
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}

func containersHandler(cfg *config.Config, st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		rows := buildContainerList(cfg, st)
		json.NewEncoder(w).Encode(rows) //nolint:errcheck
	}
}

