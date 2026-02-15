// Package widget defines the Widget interface and a Registry for managing widgets.
package widget

import (
	"context"
	"fmt"
	"html/template"
	"sync"
	"time"
)

// Widget is the interface that all data source widgets must implement.
type Widget interface {
	// Name returns the unique identifier for this widget (e.g. "weather", "rss").
	Name() string

	// Fetch retrieves fresh data and returns a pre-rendered HTML partial.
	Fetch(ctx context.Context) (*Data, error)

	// RefreshInterval returns how often this widget should be automatically refreshed.
	RefreshInterval() time.Duration

	// Enabled returns whether this widget is active.
	Enabled() bool
}

// Data holds the result of a widget fetch operation.
type Data struct {
	WidgetName string
	HTML       template.HTML
	FetchedAt  time.Time
	Error      string
	CacheTTL   time.Duration
}

// Registry manages registered widgets and their latest data.
type Registry struct {
	mu      sync.RWMutex
	widgets map[string]Widget
	order   []string
	latest  map[string]*Data
}

// NewRegistry creates an empty widget registry.
func NewRegistry() *Registry {
	return &Registry{
		widgets: make(map[string]Widget),
		latest:  make(map[string]*Data),
	}
}

// Register adds a widget to the registry.
func (r *Registry) Register(widget Widget) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := widget.Name()
	r.widgets[name] = widget
	r.order = append(r.order, name)
}

// Get returns a widget by name, or nil if not found.
func (r *Registry) Get(name string) Widget {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.widgets[name]
}

// EnabledWidgets returns the names of all enabled widgets in registration order.
func (r *Registry) EnabledWidgets() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var enabled []string
	for _, name := range r.order {
		if w, ok := r.widgets[name]; ok && w.Enabled() {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// All returns all registered widgets in registration order.
func (r *Registry) All() []Widget {
	r.mu.RLock()
	defer r.mu.RUnlock()
	widgets := make([]Widget, 0, len(r.order))
	for _, name := range r.order {
		if w, ok := r.widgets[name]; ok {
			widgets = append(widgets, w)
		}
	}
	return widgets
}

// StoreData stores the latest fetch result for a widget.
func (r *Registry) StoreData(data *Data) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.latest[data.WidgetName] = data
}

// LatestData returns the most recent fetch result for a widget.
func (r *Registry) LatestData(name string) (*Data, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, ok := r.latest[name]
	return data, ok
}

// CachedHTML returns the cached HTML for a widget, or empty if none.
func (r *Registry) CachedHTML(widgetName string) template.HTML {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if data, ok := r.latest[widgetName]; ok && data.Error == "" {
		return data.HTML
	}
	return ""
}

// WidgetData returns a map representation of a widget's data for JSON responses.
func (r *Registry) WidgetData(name string) (map[string]any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, ok := r.latest[name]
	if !ok {
		return nil, false
	}
	return map[string]any{
		"name":       data.WidgetName,
		"html":       string(data.HTML),
		"fetched_at": data.FetchedAt.Format(time.RFC3339),
		"error":      data.Error,
	}, true
}

// RefreshWidget triggers a forced refresh of a specific widget and returns the result.
func (r *Registry) RefreshWidget(ctx context.Context, name string) error {
	widget := r.Get(name)
	if widget == nil {
		return fmt.Errorf("widget %q not found", name)
	}
	if !widget.Enabled() {
		return fmt.Errorf("widget %q is disabled", name)
	}

	data, err := widget.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("fetching widget %q: %w", name, err)
	}

	r.StoreData(data)
	return nil
}
