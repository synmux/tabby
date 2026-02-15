package uptime

import (
	"context"
	"strings"
	"testing"
	"time"

	"tabby/internal/config"
)

func TestUptimeWidget(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	if widget.Name() != "uptime" {
		t.Errorf("expected name 'uptime', got %q", widget.Name())
	}

	data, err := widget.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !strings.Contains(string(data.HTML), "uptime-service") {
		t.Error("expected HTML to contain uptime-service class")
	}
	// Should contain all 5 mock services
	if strings.Count(string(data.HTML), "uptime-service") != 5 {
		t.Errorf("expected 5 services, got %d", strings.Count(string(data.HTML), "uptime-service"))
	}
}
