package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set required environment variables
	os.Setenv("CLOUDFLARE_API_TOKEN", "test-token")
	os.Setenv("CLOUDFLARE_ACCOUNT_ID", "test-account")
	os.Setenv("CLOUDFLARE_TUNNEL_ID", "test-tunnel")
	defer func() {
		os.Unsetenv("CLOUDFLARE_API_TOKEN")
		os.Unsetenv("CLOUDFLARE_ACCOUNT_ID")
		os.Unsetenv("CLOUDFLARE_TUNNEL_ID")
	}()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "test-token", cfg.CloudflareAPIToken)
	assert.Equal(t, "test-account", cfg.CloudflareAccountID)
	assert.Equal(t, "test-tunnel", cfg.CloudflareTunnelID)
	assert.Equal(t, 300*time.Second, cfg.ReconciliationInterval) // Default value
}
