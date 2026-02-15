package omi

import (
	"context"
	"strings"
	"testing"
	"time"

	"tabby/internal/config"
)

func TestOMIWidget(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	if widget.Name() != "omi" {
		t.Errorf("expected name 'omi', got %q", widget.Name())
	}

	data, err := widget.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !strings.Contains(string(data.HTML), "omi-memory") {
		t.Error("expected HTML to contain omi-memory class")
	}
}
