package calendar

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/daveio/tabby/internal/config"
)

func TestCalendarWidget(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	if widget.Name() != "calendar" {
		t.Errorf("expected name 'calendar', got %q", widget.Name())
	}

	data, err := widget.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !strings.Contains(string(data.HTML), "calendar-event") {
		t.Error("expected HTML to contain calendar-event class")
	}
}
