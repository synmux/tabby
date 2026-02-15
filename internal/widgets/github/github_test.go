package github

import (
	"context"
	"strings"
	"testing"
	"time"

	"tabby/internal/config"
)

func TestGitHubWidget(t *testing.T) {
	cfg := config.WidgetConfig{
		RefreshInterval: config.Duration{Duration: 5 * time.Second},
	}
	widget := New(cfg)

	if widget.Name() != "github" {
		t.Errorf("expected name 'github', got %q", widget.Name())
	}

	data, err := widget.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if !strings.Contains(string(data.HTML), "github-item") {
		t.Error("expected HTML to contain github-item class")
	}
}
