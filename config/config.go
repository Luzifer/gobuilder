package config

import "github.com/Luzifer/rconfig"

// Config represents the CLI / ENV config for GoBuilder-Frontend
type Config struct {
	BaseURL  string `env:"baseurl" flag:"baseurl"`
	RedisURL string `env:"redis_url" flag:"redis-url"`
	Listen   string `flag:"listen" default:":3000"`
	Port     int    `env:"PORT"` // Deprecated, only for gin

	GitHub struct {
		ClientID     string `env:"github_client_id" flag:"github-client-id"`
		ClientSecret string `env:"github_client_secret" flag:"github-client-secret"`
	}

	Papertrail struct {
		Host string `env:"papertrail_host" flag:"papertrail-host"`
		Port int    `env:"papertrail_port" flag:"papertrail-port"`
	}

	Session struct {
		AuthKey    string `env:"session_auth" flag:"session-auth"`
		EncryptKey string `env:"session_encrypt" flag:"session-encrypt"`
	}

	Pushover struct {
		APIToken string `env:"PUSHOVER_APITOKEN" flag:"pushover-token"`
	}
}

// Load collects the configuration
func Load() *Config {
	cfg := &Config{}
	rconfig.Parse(cfg)

	return cfg
}
