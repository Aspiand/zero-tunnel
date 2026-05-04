package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aspiand/zero-tunnel/internal/config"
	"github.com/aspiand/zero-tunnel/internal/engine"
	"github.com/aspiand/zero-tunnel/internal/provider"
	"github.com/aspiand/zero-tunnel/internal/watcher"
)

func main() {
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
