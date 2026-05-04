package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("Valid with Tunnel ID", func(t *testing.T) {
		os.Setenv("CLOUDFLARE_API_TOKEN", "test-token")
		os.Setenv("CLOUDFLARE_ACCOUNT_ID", "test-account")
		os.Setenv("CLOUDFLARE_TUNNEL_ID", "test-tunnel")
		os.Unsetenv("CLOUDFLARE_TUNNEL_NAME")
		defer os.Clearenv()

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "test-token", cfg.CloudflareAPIToken)
		assert.Equal(t, "test-account", cfg.CloudflareAccountID)
		assert.Equal(t, "test-tunnel", cfg.CloudflareTunnelID)
		assert.Empty(t, cfg.CloudflareTunnelName)
		assert.Equal(t, 300*time.Second, cfg.ReconciliationInterval)
	})

	t.Run("Valid with Tunnel Name", func(t *testing.T) {
		os.Setenv("CLOUDFLARE_API_TOKEN", "test-token")
		os.Setenv("CLOUDFLARE_ACCOUNT_ID", "test-account")
		os.Setenv("CLOUDFLARE_TUNNEL_NAME", "test-name")
		os.Unsetenv("CLOUDFLARE_TUNNEL_ID")
		defer os.Clearenv()

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "test-name", cfg.CloudflareTunnelName)
		assert.Empty(t, cfg.CloudflareTunnelID)
	})

	t.Run("Error if both provided", func(t *testing.T) {
		os.Setenv("CLOUDFLARE_API_TOKEN", "test-token")
		os.Setenv("CLOUDFLARE_ACCOUNT_ID", "test-account")
		os.Setenv("CLOUDFLARE_TUNNEL_ID", "test-tunnel")
		os.Setenv("CLOUDFLARE_TUNNEL_NAME", "test-name")
		defer os.Clearenv()

		_, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only one of CLOUDFLARE_TUNNEL_ID or CLOUDFLARE_TUNNEL_NAME can be provided")
	})

	t.Run("Error if neither provided", func(t *testing.T) {
		os.Setenv("CLOUDFLARE_API_TOKEN", "test-token")
		os.Setenv("CLOUDFLARE_ACCOUNT_ID", "test-account")
		os.Unsetenv("CLOUDFLARE_TUNNEL_ID")
		os.Unsetenv("CLOUDFLARE_TUNNEL_NAME")
		defer os.Clearenv()

		_, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "either CLOUDFLARE_TUNNEL_ID or CLOUDFLARE_TUNNEL_NAME must be provided")
	})
}
