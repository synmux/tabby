package rss

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/daveio/tabby/internal/config"
)

func TestRSSWidget(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	if widget.Name() != "rss" {
		t.Errorf("expected name 'rss', got %q", widget.Name())
	}

	data, err := widget.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if data.WidgetName != "rss" {
		t.Errorf("expected widget name 'rss', got %q", data.WidgetName)
	}
	if !strings.Contains(string(data.HTML), "rss-item") {
		t.Error("expected HTML to contain rss-item class")
	}
}
