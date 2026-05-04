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

		if cfg.CloudflareAPIToken == "" || cfg.CloudflareAccountID == "" || cfg.CloudflareTunnelID == "" {
			slog.Error("missing required configuration: CLOUDFLARE_API_TOKEN, CLOUDFLARE_ACCOUNT_ID, and CLOUDFLARE_TUNNEL_ID must be set")
			os.Exit(1)
		}

		w, err := watcher.New(cfg.DefaultDomain)
		if err != nil {
			return err
		}

		p := provider.New(cfg.CloudflareAPIToken, cfg.CloudflareAccountID, cfg.CloudflareTunnelID)
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
