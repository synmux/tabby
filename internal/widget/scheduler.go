package widget

import (
	"context"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"tabby/internal/cache"
	"tabby/internal/sse"
)

// Scheduler manages per-widget refresh tickers and orchestrates the
// fetch → cache → broadcast pipeline.
type Scheduler struct {
	registry *Registry
	cache    *cache.Store
	broker   *sse.Broker
	cancel   context.CancelFunc
}

// NewScheduler creates a new widget scheduler.
func NewScheduler(registry *Registry, cacheStore *cache.Store, broker *sse.Broker) *Scheduler {
	return &Scheduler{
		registry: registry,
		cache:    cacheStore,
		broker:   broker,
	}
}

// Start launches a goroutine per enabled widget that fetches on its refresh interval.
// It also performs an initial fetch of all widgets.
func (s *Scheduler) Start(parentCtx context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)
	s.cancel = cancel

	// Load cached data into registry on startup
	s.loadCachedData()

	// Initial fetch of all widgets
	s.fetchAndBroadcastAll(ctx)

	// Start per-widget tickers
	for _, widget := range s.registry.All() {
		if !widget.Enabled() {
			continue
		}
		go s.runWidgetTicker(ctx, widget)
	}
}

// Stop cancels all scheduler goroutines.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Scheduler) loadCachedData() {
	for _, name := range s.registry.EnabledWidgets() {
		entry, err := s.cache.Get(name)
		if err != nil {
			slog.Warn("failed to load cached data", "widget", name, "error", err)
			continue
		}
		if entry != nil {
			s.registry.StoreData(&Data{
				WidgetName: entry.WidgetName,
				HTML:       template.HTML(entry.HTML),
				FetchedAt:  entry.FetchedAt,
				Error:      entry.Error,
			})
			slog.Debug("loaded cached data", "widget", name)
		}
	}
}

func (s *Scheduler) fetchAndBroadcastAll(ctx context.Context) {
	results := FetchAll(ctx, s.registry)
	for _, result := range results {
		if result.Data == nil {
			continue
		}
		s.processResult(result.Data)
	}
}

func (s *Scheduler) runWidgetTicker(ctx context.Context, widget Widget) {
	ticker := time.NewTicker(widget.RefreshInterval())
	defer ticker.Stop()

	slog.Info("widget ticker started",
		"widget", widget.Name(),
		"interval", widget.RefreshInterval().String(),
	)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			data, err := widget.Fetch(ctx)
			if err != nil {
				slog.Warn("widget fetch failed",
					"widget", widget.Name(),
					"error", err,
				)
				continue
			}
			s.processResult(data)
		}
	}
}

func (s *Scheduler) processResult(data *Data) {
	// Store in registry (in-memory)
	s.registry.StoreData(data)

	// Persist to SQLite cache
	cacheTTL := data.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 10 * time.Minute
	}

	cacheEntry := &cache.Entry{
		WidgetName: data.WidgetName,
		HTML:       string(data.HTML),
		FetchedAt:  data.FetchedAt,
		ExpiresAt:  data.FetchedAt.Add(cacheTTL),
		Error:      data.Error,
	}
	if err := s.cache.Set(cacheEntry); err != nil {
		slog.Warn("failed to cache widget data", "widget", data.WidgetName, "error", err)
	}

	// Broadcast via SSE (only if no error)
	if data.Error == "" {
		// SSE data must be single-line; replace newlines with a separate data: line
		htmlStr := string(data.HTML)
		// Collapse the HTML into SSE-safe format: join lines with empty space
		sseData := strings.ReplaceAll(htmlStr, "\n", "\ndata: ")

		s.broker.Broadcast(sse.Event{
			Name: data.WidgetName,
			Data: sseData,
		})
	}
}

// RefreshWidget forces an immediate refresh of a single widget and broadcasts the result.
func (s *Scheduler) RefreshWidget(ctx context.Context, name string) error {
	if err := s.registry.RefreshWidget(ctx, name); err != nil {
		return err
	}

	data, ok := s.registry.LatestData(name)
	if ok {
		s.processResult(data)
	}
	return nil
}

// EnabledWidgets returns the names of all enabled widgets.
func (s *Scheduler) EnabledWidgets() []string {
	return s.registry.EnabledWidgets()
}

// WidgetData returns a map representation of a widget's data.
func (s *Scheduler) WidgetData(name string) (map[string]any, bool) {
	return s.registry.WidgetData(name)
}
