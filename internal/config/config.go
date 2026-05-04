package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	CloudflareAPIToken     string        `mapstructure:"CLOUDFLARE_API_TOKEN"`
	CloudflareAccountID    string        `mapstructure:"CLOUDFLARE_ACCOUNT_ID"`
	CloudflareTunnelID     string        `mapstructure:"CLOUDFLARE_TUNNEL_ID"`
	DefaultDomain          string        `mapstructure:"ZERO_TUNNEL_DEFAULT_DOMAIN"`
	ReconciliationInterval time.Duration `mapstructure:"ZERO_TUNNEL_INTERVAL"`
}

func Load() (*Config, error) {
	viper.SetDefault("ZERO_TUNNEL_INTERVAL", 300*time.Second)

	viper.AutomaticEnv()

	_ = viper.BindEnv("CLOUDFLARE_API_TOKEN")
	_ = viper.BindEnv("CLOUDFLARE_ACCOUNT_ID")
	_ = viper.BindEnv("CLOUDFLARE_TUNNEL_ID")
	_ = viper.BindEnv("ZERO_TUNNEL_DEFAULT_DOMAIN")
	_ = viper.BindEnv("ZERO_TUNNEL_INTERVAL")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
