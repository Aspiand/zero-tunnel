package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aspiand/zero-tunnel/internal/config"
	"github.com/aspiand/zero-tunnel/internal/engine"
	"github.com/aspiand/zero-tunnel/internal/provider"
	"github.com/aspiand/zero-tunnel/internal/watcher"
)

var (
	Version   = "(dev)"
	Commit    = "(dev)"
	BuildDate = "(dev)"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("zero-tunnel %s\n", Version)
			fmt.Printf("commit: %s\n", Commit)
			fmt.Printf("build date: %s\n", BuildDate)
			return
		}
	}

	slog.Info(
		"starting zero-tunnel",
		"version", Version,
		"commit", Commit,
		"build_date", BuildDate,
	)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	w, err := watcher.New(cfg.DefaultDomain)
	if err != nil {
		slog.Error("failed to initialize watcher", "error", err)
		os.Exit(1)
	}

	p := provider.New(cfg.CloudflareAPIToken, cfg.CloudflareAccountID, cfg.CloudflareTunnelID)

	// Lookup tunnel ID by name if provided
	if cfg.CloudflareTunnelName != "" {
		tunnelID, err := p.LookupTunnelID(context.Background(), cfg.CloudflareTunnelName)
		if err != nil {
			slog.Error("failed to lookup tunnel ID", "name", cfg.CloudflareTunnelName, "error", err)
			os.Exit(1)
		}
		slog.Info("resolved tunnel ID", "name", cfg.CloudflareTunnelName, "id", tunnelID)
		p.SetTunnelID(tunnelID)
	}

	eng := engine.New(w, p, cfg.ReconciliationInterval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := eng.Run(ctx); err != nil {
		slog.Error("engine execution failed", "error", err)
		os.Exit(1)
	}
}
