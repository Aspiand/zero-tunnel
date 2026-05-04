package cmd

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
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zero-tunnel",
	Short: "Automated Cloudflare Tunnel routing for Docker",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		w, err := watcher.New(cfg.DefaultDomain)
		if err != nil {
			return err
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

		return eng.Run(ctx)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("execution failed", "error", err)
		os.Exit(1)
	}
}
