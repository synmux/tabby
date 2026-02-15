// Package config handles TOML configuration parsing for the Tabby server.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config is the top-level configuration structure parsed from tabby.toml.
type Config struct {
	Server  ServerConfig            `toml:"server"`
	Widgets map[string]WidgetConfig `toml:"widgets"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	TLS      bool   `toml:"tls"`
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
}

// WidgetConfig holds per-widget settings.
type WidgetConfig struct {
	Enabled         *bool             `toml:"enabled"`
	RefreshInterval Duration          `toml:"refresh_interval"`
	Settings        map[string]string `toml:"settings"`
}

// Duration wraps time.Duration to support TOML string parsing (e.g. "30s", "5m").
type Duration struct {
	time.Duration
}

// UnmarshalText parses a duration string like "30s" or "5m" from TOML.
func (d *Duration) UnmarshalText(text []byte) error {
	parsed, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", string(text), err)
	}
	d.Duration = parsed
	return nil
}

// IsEnabled returns whether the widget is enabled, defaulting to true.
func (wc WidgetConfig) IsEnabled() bool {
	if wc.Enabled == nil {
		return true
	}
	return *wc.Enabled
}

// GetSetting returns a widget setting value, or the fallback if not set.
func (wc WidgetConfig) GetSetting(key, fallback string) string {
	if value, ok := wc.Settings[key]; ok {
		return value
	}
	return fallback
}

// DefaultConfig returns a Config with sensible defaults for all settings.
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Widgets: map[string]WidgetConfig{
			"weather":  {RefreshInterval: Duration{5 * time.Minute}},
			"rss":      {RefreshInterval: Duration{10 * time.Minute}},
			"uptime":   {RefreshInterval: Duration{30 * time.Second}},
			"github":   {RefreshInterval: Duration{2 * time.Minute}},
			"linear":   {RefreshInterval: Duration{2 * time.Minute}},
			"notion":   {RefreshInterval: Duration{3 * time.Minute}},
			"omi":      {RefreshInterval: Duration{5 * time.Minute}},
			"calendar": {RefreshInterval: Duration{1 * time.Minute}},
		},
	}
}

// Load reads and parses a TOML configuration file, merging with defaults.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return &cfg, nil
}

// Address returns the host:port string for the server to listen on.
func (sc ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", sc.Host, sc.Port)
}
