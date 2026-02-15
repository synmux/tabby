package weather

import (
	"context"
	"testing"
	"time"

	"tabby/internal/config"
)

func TestWeatherWidget(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	if widget.Name() != "weather" {
		t.Errorf("expected name 'weather', got %q", widget.Name())
	}
	if !widget.Enabled() {
		t.Error("expected widget to be enabled by default")
	}
	if widget.RefreshInterval() != 5*time.Second {
		t.Errorf("expected 5s refresh, got %v", widget.RefreshInterval())
	}

	data, err := widget.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if data.WidgetName != "weather" {
		t.Errorf("expected widget name 'weather', got %q", data.WidgetName)
	}
	if data.HTML == "" {
		t.Error("expected non-empty HTML")
	}
	if data.FetchedAt.IsZero() {
		t.Error("expected non-zero FetchedAt")
	}
}

func TestWeatherWidgetDynamicData(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	// Fetch multiple times and verify data varies (not identical)
	results := make(map[string]bool)
	for range 10 {
		data, err := widget.Fetch(context.Background())
		if err != nil {
			t.Fatalf("Fetch failed: %v", err)
		}
		results[string(data.HTML)] = true
	}

	// With random data, we should see more than 1 unique result in 10 fetches
	if len(results) < 2 {
		t.Error("expected dynamic data across fetches, got identical results")
	}
}
