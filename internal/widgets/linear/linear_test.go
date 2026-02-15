package linear

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/daveio/tabby/internal/config"
)

func TestLinearWidget(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	if widget.Name() != "linear" {
		t.Errorf("expected name 'linear', got %q", widget.Name())
	}

	data, err := widget.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !strings.Contains(string(data.HTML), "linear-ticket") {
		t.Error("expected HTML to contain linear-ticket class")
	}
}
