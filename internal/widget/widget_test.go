package widget

import (
	"context"
	"html/template"
	"testing"
	"time"
)

// mockWidget implements the Widget interface for testing.
type mockWidget struct {
	name     string
	enabled  bool
	interval time.Duration
	html     string
	err      error
}

func (m *mockWidget) Name() string                  { return m.name }
func (m *mockWidget) Enabled() bool                 { return m.enabled }
func (m *mockWidget) RefreshInterval() time.Duration { return m.interval }
func (m *mockWidget) Fetch(_ context.Context) (*Data, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &Data{
		WidgetName: m.name,
		HTML:       template.HTML(m.html),
		FetchedAt:  time.Now(),
		CacheTTL:   m.interval,
	}, nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	w := &mockWidget{name: "test", enabled: true, interval: time.Minute, html: "<div>test</div>"}
	registry.Register(w)

	got := registry.Get("test")
	if got == nil {
		t.Fatal("expected widget, got nil")
	}
	if got.Name() != "test" {
		t.Errorf("expected name 'test', got %q", got.Name())
	}
}

func TestRegistryGetMissing(t *testing.T) {
	registry := NewRegistry()

	got := registry.Get("nonexistent")
	if got != nil {
		t.Error("expected nil for missing widget")
	}
}

func TestRegistryEnabledWidgets(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&mockWidget{name: "enabled1", enabled: true})
	registry.Register(&mockWidget{name: "disabled", enabled: false})
	registry.Register(&mockWidget{name: "enabled2", enabled: true})

	enabled := registry.EnabledWidgets()
	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled widgets, got %d", len(enabled))
	}
	if enabled[0] != "enabled1" || enabled[1] != "enabled2" {
		t.Errorf("unexpected order: %v", enabled)
	}
}

func TestRegistryAll(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&mockWidget{name: "a", enabled: true})
	registry.Register(&mockWidget{name: "b", enabled: false})

	all := registry.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 widgets, got %d", len(all))
	}
}

func TestRegistryStoreAndRetrieveData(t *testing.T) {
	registry := NewRegistry()

	data := &Data{
		WidgetName: "weather",
		HTML:       "<div>sunny</div>",
		FetchedAt:  time.Now(),
	}

	registry.StoreData(data)

	got, ok := registry.LatestData("weather")
	if !ok {
		t.Fatal("expected data to exist")
	}
	if got.WidgetName != "weather" {
		t.Errorf("expected 'weather', got %q", got.WidgetName)
	}
}

func TestRegistryCachedHTML(t *testing.T) {
	registry := NewRegistry()

	// No data — should return empty
	if html := registry.CachedHTML("weather"); html != "" {
		t.Errorf("expected empty HTML, got %q", html)
	}

	// With data
	registry.StoreData(&Data{
		WidgetName: "weather",
		HTML:       "<div>sunny</div>",
		FetchedAt:  time.Now(),
	})
	if html := registry.CachedHTML("weather"); html != "<div>sunny</div>" {
		t.Errorf("expected cached HTML, got %q", html)
	}

	// With error — should return empty
	registry.StoreData(&Data{
		WidgetName: "broken",
		HTML:       "<div>data</div>",
		Error:      "something failed",
	})
	if html := registry.CachedHTML("broken"); html != "" {
		t.Errorf("expected empty HTML for errored widget, got %q", html)
	}
}

func TestRegistryWidgetData(t *testing.T) {
	registry := NewRegistry()

	_, ok := registry.WidgetData("missing")
	if ok {
		t.Error("expected false for missing widget data")
	}

	registry.StoreData(&Data{
		WidgetName: "weather",
		HTML:       "<div>data</div>",
		FetchedAt:  time.Now(),
	})

	data, ok := registry.WidgetData("weather")
	if !ok {
		t.Fatal("expected widget data to exist")
	}
	if data["name"] != "weather" {
		t.Errorf("expected name 'weather', got %v", data["name"])
	}
}
