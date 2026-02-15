package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Address() != "localhost:8080" {
		t.Errorf("expected address localhost:8080, got %s", cfg.Server.Address())
	}
	if len(cfg.Widgets) != 8 {
		t.Errorf("expected 8 default widgets, got %d", len(cfg.Widgets))
	}
	if cfg.Widgets["weather"].RefreshInterval.Duration != 5*time.Second {
		t.Errorf("expected weather default 5s, got %v", cfg.Widgets["weather"].RefreshInterval.Duration)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/tabby.toml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	// Should return defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadValidTOML(t *testing.T) {
	content := `
[server]
host = "0.0.0.0"
port = 9090

[widgets.weather]
enabled = true
refresh_interval = "30s"

[widgets.rss]
enabled = false
refresh_interval = "15m"
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "tabby.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}

	weatherCfg := cfg.Widgets["weather"]
	if !weatherCfg.IsEnabled() {
		t.Error("expected weather to be enabled")
	}
	if weatherCfg.RefreshInterval.Duration != 30*time.Second {
		t.Errorf("expected 30s refresh, got %v", weatherCfg.RefreshInterval.Duration)
	}

	rssCfg := cfg.Widgets["rss"]
	if rssCfg.IsEnabled() {
		t.Error("expected rss to be disabled")
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.toml")
	if err := os.WriteFile(path, []byte("{{{{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestWidgetConfigDefaults(t *testing.T) {
	// nil Enabled should default to true
	wc := WidgetConfig{}
	if !wc.IsEnabled() {
		t.Error("expected nil Enabled to default to true")
	}

	// Explicit false
	falseBool := false
	wc.Enabled = &falseBool
	if wc.IsEnabled() {
		t.Error("expected explicit false to be disabled")
	}
}

func TestWidgetConfigGetSetting(t *testing.T) {
	wc := WidgetConfig{
		Settings: map[string]string{
			"api_key": "secret123",
		},
	}

	if got := wc.GetSetting("api_key", ""); got != "secret123" {
		t.Errorf("expected secret123, got %s", got)
	}
	if got := wc.GetSetting("missing", "fallback"); got != "fallback" {
		t.Errorf("expected fallback, got %s", got)
	}
}

func TestDurationUnmarshalText(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"30s", 30 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"invalid", 0, true},
	}

	for _, testCase := range tests {
		var duration Duration
		err := duration.UnmarshalText([]byte(testCase.input))
		if testCase.wantErr {
			if err == nil {
				t.Errorf("expected error for %q", testCase.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error for %q: %v", testCase.input, err)
			continue
		}
		if duration.Duration != testCase.expected {
			t.Errorf("for %q: expected %v, got %v", testCase.input, testCase.expected, duration.Duration)
		}
	}
}
