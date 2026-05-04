package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	CloudflareAPIToken     string        `mapstructure:"CLOUDFLARE_API_TOKEN"`
	CloudflareAccountID    string        `mapstructure:"CLOUDFLARE_ACCOUNT_ID"`
	CloudflareTunnelID     string        `mapstructure:"CLOUDFLARE_TUNNEL_ID"`
	CloudflareTunnelName   string        `mapstructure:"CLOUDFLARE_TUNNEL_NAME"`
	DefaultDomain          string        `mapstructure:"ZERO_TUNNEL_DEFAULT_DOMAIN"`
	ReconciliationInterval time.Duration `mapstructure:"ZERO_TUNNEL_INTERVAL"`
}

func (c *Config) Validate() error {
	if c.CloudflareAPIToken == "" {
		return fmt.Errorf("CLOUDFLARE_API_TOKEN is required")
	}
	if c.CloudflareAccountID == "" {
		return fmt.Errorf("CLOUDFLARE_ACCOUNT_ID is required")
	}
	if c.CloudflareTunnelID == "" && c.CloudflareTunnelName == "" {
		return fmt.Errorf("either CLOUDFLARE_TUNNEL_ID or CLOUDFLARE_TUNNEL_NAME must be provided")
	}
	if c.CloudflareTunnelID != "" && c.CloudflareTunnelName != "" {
		return fmt.Errorf("only one of CLOUDFLARE_TUNNEL_ID or CLOUDFLARE_TUNNEL_NAME can be provided")
	}
	return nil
}

func Load() (*Config, error) {
	viper.SetDefault("ZERO_TUNNEL_INTERVAL", 300*time.Second)

	viper.AutomaticEnv()

	_ = viper.BindEnv("CLOUDFLARE_API_TOKEN")
	_ = viper.BindEnv("CLOUDFLARE_ACCOUNT_ID")
	_ = viper.BindEnv("CLOUDFLARE_TUNNEL_ID")
	_ = viper.BindEnv("CLOUDFLARE_TUNNEL_NAME")
	_ = viper.BindEnv("ZERO_TUNNEL_DEFAULT_DOMAIN")
	_ = viper.BindEnv("ZERO_TUNNEL_INTERVAL")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
