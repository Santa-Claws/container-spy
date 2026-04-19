package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmac/container-spy/internal/config"
	dockerpkg "github.com/tmac/container-spy/internal/docker"
	"github.com/tmac/container-spy/internal/store"
	"github.com/tmac/container-spy/internal/tui"
	"github.com/tmac/container-spy/internal/tui/pages"
	"github.com/tmac/container-spy/internal/web"
)

func main() {
	mode := os.Getenv("CONTAINER_SPY_MODE")
	if mode == "" {
		mode = "tui"
	}
	cfgPath := os.Getenv("CONTAINER_SPY_CONFIG")
	if cfgPath == "" {
		cfgPath = config.DefaultPath
	}

	// In TUI mode, redirect logs to a file so they don't corrupt the terminal.
	if mode == "tui" {
		logFile, err := os.OpenFile("/tmp/container-spy.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			slog.SetDefault(slog.New(slog.NewTextHandler(logFile, nil)))
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Write SSH config before creating any Docker clients.
	if err := dockerpkg.GenerateSSHConfig(cfg.Servers); err != nil {
		slog.Warn("Could not write SSH config", "err", err)
	}

	st := store.New()
	pollerMgr := dockerpkg.NewManager(st)
	pollerMgr.Sync(cfg.Servers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
		pollerMgr.StopAll()
	}()

	switch mode {
	case "web":
		runWeb(ctx, cfg, cfgPath, st, pollerMgr)
	default:
		runTUI(ctx, cfg, cfgPath, st, pollerMgr)
	}
}

func runWeb(ctx context.Context, cfg *config.Config, cfgPath string, st *store.Store, pollerMgr *dockerpkg.Manager) {
	addr := os.Getenv("CONTAINER_SPY_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	slog.Info("Starting WebUI", "addr", addr)
	srv := web.NewServer(addr, cfg, st)
	if err := srv.Start(ctx); err != nil {
		slog.Error("Web server error", "err", err)
		os.Exit(1)
	}
}

func runTUI(ctx context.Context, cfg *config.Config, cfgPath string, st *store.Store, pollerMgr *dockerpkg.Manager) {
	keys := tui.DefaultKeyMap()

	tuiPages := map[tui.PageID]tui.Page{
		tui.PageDashboard: func() tui.Page {
			d := pages.NewDashboard(cfg, st, keys)
			return &d
		}(),
		tui.PageServers: func() tui.Page {
			s := pages.NewServers(cfg, cfgPath, keys)
			return &s
		}(),
		tui.PageGroups: func() tui.Page {
			g := pages.NewGroups(cfg, cfgPath, st, keys)
			return &g
		}(),
	}

	model := tui.New(cfg, cfgPath, st, tuiPages)
	model.OnSync = func() {
		dockerpkg.GenerateSSHConfig(cfg.Servers) //nolint:errcheck
		pollerMgr.Sync(cfg.Servers)
	}

	prog := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Give pollers a reference so they can send RefreshMsg to the TUI.
	pollerMgr.SetProgram(prog)

	if _, err := prog.Run(); err != nil {
		slog.Error("TUI error", "err", err)
		os.Exit(1)
	}
}
