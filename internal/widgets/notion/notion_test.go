package notion

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/daveio/tabby/internal/config"
)

func TestNotionWidget(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	if widget.Name() != "notion" {
		t.Errorf("expected name 'notion', got %q", widget.Name())
	}

	data, err := widget.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !strings.Contains(string(data.HTML), "notion-page") {
		t.Error("expected HTML to contain notion-page class")
	}
}
